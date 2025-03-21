package modules

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

// +kubebuilder:object:generate=true
type KernelModules struct {
	Modules []string `json:"modules,omitempty"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state,omitempty"`
}

// IsPresent method checks if the module is present
func (k KernelModules) IsPresent() bool {
	if len(k.Modules) != 0 && k.State == "present" {
		return true
	}
	return false
}

type KernelModule = string

type KernelModuleConfig struct {
	KernelModules
	logger logr.Logger
	// This file is for loading the kernel modules at boot
	// by systemd-modules-load
	filePath string
}

func NewKernelModuleConfig(modules KernelModules, log logr.Logger) KernelModuleConfig {
	return KernelModuleConfig{
		KernelModules: modules,
		logger:        log,
		filePath:      "/etc/modules-load.d/nco.conf",
	}
}

func (c KernelModuleConfig) Reconcile() error {
	moduleError := ModuleError{"kernelModules", nil}
	if c.State == "present" {
		c.logger.V(1).Info("applying module")
		if err := c.applyModule(); err != nil {
			moduleError.error = err
			return moduleError
		}
		c.logger.V(1).Info("module applied")
	} else if c.State == "absent" {
		c.logger.V(1).Info("removing module")
		if err := c.removeModule(); err != nil {
			moduleError.error = err
			return moduleError
		}
		c.logger.V(1).Info("module removed")
	}

	return nil
}

func (c KernelModuleConfig) applyModule() error {
	isCurrent, err := c.checkCurrentConfig()
	if err != nil {
		return fmt.Errorf("failed to check current config: %w", err)
	}

	if isCurrent {
		// do nothing
		return nil
	}

	err = writeFile(c.filePath, strings.Join(c.Modules, "\n"))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	for _, module := range c.Modules {
		cmd := exec.Command("modprobe", module)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run modprobe: %w", err)
		}
	}

	return nil
}

func (c KernelModuleConfig) removeModule() error {
	err := os.Remove(c.filePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// Modules shouldn't be unloaded, next host reboot should fix
	// the inconsistency
	c.logger.V(1).Info("finished cleaning up")

	return nil
}

func (c KernelModuleConfig) checkCurrentConfig() (bool, error) {
	isFileEqual, err := checkFileContents(c.filePath, strings.Join(c.Modules, "\n"))
	if err != nil {
		return false, err
	}

	if !isFileEqual {
		return false, nil
	}

	for _, module := range c.Modules {
		if !isModuleActive(module) {
			return false, nil
		}
	}

	return true, nil
}

func isModuleActive(moduleName string) bool {
	cmd := exec.Command("lsmod", moduleName)
	err := cmd.Run()
	return err == nil
}
