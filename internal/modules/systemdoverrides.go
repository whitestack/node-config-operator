package modules

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
)

const (
	overrideBasePath = "/host/etc/systemd/system"
	overridePrevName = "90-nco-override.conf"
	overrideHeader   = "# FILE MANAGED BY NCO - CHANGES TO THIS FILE WILL BE OVERWRITTEN"
)

const (
	SERVICE_TYPE = "service"
	SLICE_TYPE   = "slice"
)

// +kubebuilder:object:generate=true
type SystemdOverrides struct {
	Overrides []SystemdOverride `json:"overrides,omitempty"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state,omitempty"`
}

// IsPresent method checks if the module is present
func (s SystemdOverrides) IsPresent() bool {
	if len(s.Overrides) != 0 && s.State == "present" {
		return true
	}
	return false
}

// +kubebuilder:object:generate=true
type SystemdOverride struct {
	// Name of unit to override, must have service or slice suffix
	Name string `json:"name"`
	// Contents of file
	File string `json:"file"`
	// Priority to set for these overrides (default: 50)
	// +kubebuilder:validation:Maximum:=99
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=50
	// +optional
	Priority *int `json:"priority,omitempty"`
}

type systemdOverride struct {
	unitName    string
	unitType    string
	fileContent string
	priority    *int
}

type SystemdOverrideConfig struct {
	overrides    []systemdOverride
	state        string
	logger       logr.Logger
	resourceName string
}

func NewSystemdOverrideConfig(overrides SystemdOverrides, logger logr.Logger, name string) SystemdOverrideConfig {
	ov := make([]systemdOverride, len(overrides.Overrides))

	for i, override := range overrides.Overrides {
		var unitType string
		if strings.HasSuffix(override.Name, ".service") {
			unitType = SERVICE_TYPE
		} else if strings.HasSuffix(override.Name, ".slice") {
			unitType = SLICE_TYPE
		} else {
			logger.Info("Warning: unit type not supported", "unitName", override.Name)
			continue
		}

		ov[i] = systemdOverride{
			unitName:    override.Name,
			unitType:    unitType,
			fileContent: override.File,
			priority:    override.Priority,
		}
	}

	return SystemdOverrideConfig{
		overrides:    ov,
		state:        overrides.State,
		logger:       logger,
		resourceName: name,
	}
}

func (s SystemdOverrideConfig) Reconcile() error {
	hostFsEnabled := os.Getenv("HOSTFS_ENABLED")
	if hostFsEnabled != "true" {
		err := errors.New("HOSTFS_ENABLED is set to false")
		s.logger.Error(err, "module needs chroot to work, set HOSTFS_ENABLED to true")
		return nil
	}

	moduleError := ModuleError{"systemdOverrides", nil}
	if s.state == "present" {
		s.logger.V(1).Info("applying module")
		if err := s.applyModule(); err != nil {
			moduleError.error = err
			return moduleError
		}
		s.logger.V(1).Info("module applied")
	} else if s.state == "absent" {
		s.logger.V(1).Info("removing module")
		if err := s.removeModule(); err != nil {
			moduleError.error = err
			return moduleError
		}
		s.logger.V(1).Info("module removed")
	}

	return nil
}

func (s SystemdOverrideConfig) applyModule() error {
	needsRestart := make([]bool, len(s.overrides))
	for i, override := range s.overrides {
		folderPath := fmt.Sprintf("%s/%s.d", overrideBasePath, override.unitName)

		// delete previous file as it's not needed anymore
		prevFileName := fmt.Sprintf("%s/%s", folderPath, overridePrevName)
		if err := deleteFileIfExists(prevFileName); err != nil {
			return fmt.Errorf("failed to delete prevFileName: %w", err)
		}

		// Override files location is in
		// `/etc/systemd/system/<unit-name>.d/<override-file>`, for example
		// `/etc/systemd/system/getty@tty2.service.d/override.conf`so we build
		// the complete file path with the unit information
		overrideName := fmt.Sprintf("%d-nco-%s-override.conf", *override.priority, s.resourceName)
		filePath := fmt.Sprintf("%s/%s", folderPath, overrideName)
		content := overrideHeader + "\n" + override.fileContent

		if err := checkOrCreateDirectory(folderPath); err != nil {
			return fmt.Errorf("failed to create unit override folder: %w", err)
		}

		isFileCorrect, err := checkFileContents(filePath, content)
		if err != nil {
			return fmt.Errorf("failed to check file: %w", err)
		}

		if !isFileCorrect {
			if err := writeFile(filePath, content); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			needsRestart[i] = true
		}
	}

	// Reload systemd configuration
	_, err := execChroot("systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("failed to reload daemon: %w", err)
	}

	for i, override := range s.overrides {
		// Services will be restarted to load its new configuration when the
		// override has changed.
		// The slice's override will apply to new processes or to all processes
		// on system boot
		if override.unitType != SERVICE_TYPE || !needsRestart[i] {
			continue
		}

		_, err := execChroot("systemctl", "restart", override.unitName)
		if err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}
	}
	return nil
}

func (s SystemdOverrideConfig) removeModule() error {
	for _, override := range s.overrides {
		folderPath := fmt.Sprintf("%s/%s.d", overrideBasePath, override.unitName)

		// delete previous file as it's not needed anymore
		prevFileName := fmt.Sprintf("%s/%s", folderPath, overridePrevName)
		if err := deleteFileIfExists(prevFileName); err != nil {
			return fmt.Errorf("failed to delete prevFileName: %w", err)
		}

		overrideName := fmt.Sprintf("%d-nco-%s-override.conf", *override.priority, s.resourceName)
		filePath := fmt.Sprintf("%s/%s", folderPath, overrideName)

		if err := deleteFileIfExists(filePath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// Reload systemd configuration
	_, err := execChroot("systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("failed to reload daemon: %w", err)
	}

	for _, override := range s.overrides {
		if override.unitType != SERVICE_TYPE {
			continue
		}

		_, err := execChroot("systemctl", "restart", override.unitName)
		if err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}
	}

	return nil
}
