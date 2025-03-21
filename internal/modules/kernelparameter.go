package modules

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

// +kubebuilder:object:generate=true
type KernelParameters struct {
	Parameters []KernelParameterKV `json:"parameters,omitempty"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state,omitempty"`
	// Priority to set for these parameters (default: 50)
	// +kubebuilder:validation:Maximum:=99
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=50
	// +optional
	Priority *int `json:"priority,omitempty"`
}

// IsPresent method checks if the module is present
func (k KernelParameters) IsPresent() bool {
	if len(k.Parameters) != 0 && k.State == "present" {
		return true
	}
	return false
}

type KernelParameterKV struct {
	// Name of the kernel parameter (e.g. fs.file-max)
	Name string `json:"name,omitempty"`
	// Desired value of the kernel parameter
	Value string `json:"value,omitempty"`
}

type KernelParameterConfig struct {
	KernelParameters
	logger       logr.Logger
	filePath     string
	prevFilePath string
}

func NewKernelParameterConfig(configs KernelParameters, log logr.Logger, name string) KernelParameterConfig {
	folder := "/etc/sysctl.d/"

	filePath := fmt.Sprintf("%s/%d-nco-%s.conf", folder, *configs.Priority, name)

	return KernelParameterConfig{
		KernelParameters: configs,
		logger:           log,
		filePath:         filePath,
		prevFilePath:     "/etc/sysctl.d/99-nco.conf",
	}
}

func (c KernelParameterConfig) Reconcile() error {
	moduleError := ModuleError{"kernelParameter", nil}
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

func (c KernelParameterConfig) applyModule() error {
	// delete prevFilePath as it's not needed anymore
	// as we use a different file for each NCO resource
	if err := deleteFileIfExists(c.prevFilePath); err != nil {
		return fmt.Errorf("failed to remove prevFilePath: %w", err)
	}

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

	cmd := exec.Command("sysctl", "-p", c.filePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error applying sysctl config: %s", err)
	}

	return nil
}

func (c KernelParameterConfig) removeModule() error {
	// delete prevFilePath as it's not needed anymore
	// as we use a different file for each NCO resource
	if err := deleteFileIfExists(c.prevFilePath); err != nil {
		return fmt.Errorf("failed to remove prevFilePath: %w", err)
	}

	// Attempt to remove the file
	if err := deleteFileIfExists(c.filePath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	// reload sysctl configuration
	cmd := exec.Command("sysctl", "-p")
	err := cmd.Run()
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
