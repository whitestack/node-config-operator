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
type KernelParameters struct {
	Parameters []KernelParameterKV `json:"parameters"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state"`
}

type KernelParameterKV struct {
	// Name of the kernel parameter (e.g. fs.file-max)
	Name string `json:"name,omitempty"`
	// Desired value of the kernel parameter
	Value string `json:"value,omitempty"`
}

type KernelParameterConfig struct {
	KernelParameters
	logger   logr.Logger
	filePath string
}

func NewKernelParameterConfig(configs KernelParameters, log logr.Logger) KernelParameterConfig {
	return KernelParameterConfig{
		KernelParameters: configs,
		logger:           log,
		filePath:         "/etc/sysctl.d/99-nco.conf",
	}
}

func (c KernelParameterConfig) Reconcile() error {
	if c.State == "present" {
		c.logger.V(1).Info("applying module")
		if err := c.applyModule(); err != nil {
			return fmt.Errorf("failed to apply module: %w", err)
		}
		c.logger.V(1).Info("module applied")
	} else if c.State == "absent" {
		c.logger.V(1).Info("removing module")
		if err := c.applyModule(); err != nil {
			return fmt.Errorf("failed to remove module: %w", err)
		}
		c.logger.V(1).Info("module removed")
	}

	return nil
}

func (c KernelParameterConfig) applyModule() error {
	// check current configuration
	newParameters := make([]string, len(c.Parameters))
	for i, parameters := range c.Parameters {
		newParameters[i] = parameters.Name + " = " + parameters.Value
	}

	isCurrent, err := c.checkCurrentConfig(newParameters)
	if err != nil {
		return fmt.Errorf("failed to check current configuration: %w", err)
	}

	if isCurrent {
		// do nothing
		return nil
	}

	// generate a config file from all configs
	err = writeFile(c.filePath, strings.Join(newParameters, "\n"))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// apply /etc/sysctl.d/99-nco.conf configuration
	cmd := exec.Command("sysctl", "-p", c.filePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error applying sysctl config: %s", err)
	}

	return nil
}

func (c KernelParameterConfig) removeModule() error {
	// Attempt to remove the file
	err := os.Remove(c.filePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	// reload sysctl configuration
	cmd := exec.Command("sysctl", "-p")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error applying sysctl config: %s", err)
	}
	c.logger.V(1).Info("finished cleaning up")
	return nil
}

func (c KernelParameterConfig) checkCurrentConfig(newConfigLines []string) (bool, error) {
	isFileEqual, err := checkFileContents(c.filePath, strings.Join(newConfigLines, "\n"))
	if err != nil {
		return false, err
	}

	if !isFileEqual {
		return false, nil
	}

	// check if each new config is currently applied
	for _, config := range c.Parameters {
		// check and compare value
		isEqual, err := isSysctlEqual(config.Name, config.Value)
		if err != nil {
			// This parameter is incorrect
			return false, err
		} else if !isEqual {
			return false, nil
		}
	}
	return true, nil
}

func isSysctlEqual(parameter string, desiredValue string) (bool, error) {
	suffix := strings.ReplaceAll(parameter, ".", "/")
	filePath := "/proc/sys/" + suffix

	// Read the content of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("Could not read sysctl %s : %w", parameter, err)
	}

	// Convert content to string
	currentValue := strings.TrimSpace(string(content))

	return currentValue == desiredValue, nil
}
