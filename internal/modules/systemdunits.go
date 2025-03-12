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
type SystemdUnits struct {
	Units []SystemdUnit `json:"units,omitempty"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state,omitempty"`
}

// IsPresent method checks if the module is present
func (s SystemdUnits) IsPresent() bool {
	if len(s.Units) != 0 && s.State == "present" {
		return true
	}
	return false
}

type SystemdUnit struct {
	// Name of the service. A "nco" prefix will be appended
	Name string `json:"name"`
	// Contents of the systemd unit
	File string `json:"file"`
}

type systemdUnit struct {
	absPath      string
	serviceName  string
	fileContents string
}

type SystemdUnitConfig struct {
	units  []systemdUnit
	state  string
	logger logr.Logger
}

const systemdPath = "/host/etc/systemd/system"

func NewSystemdUnitConfig(units SystemdUnits, logger logr.Logger) SystemdUnitConfig {
	s := make([]systemdUnit, len(units.Units))

	for i, unit := range units.Units {
		serviceName := "nco-" + unit.Name

		if strings.HasSuffix(serviceName, ".timer") {
			logger.Info("Warning: service shouldn't be of type timer", "serviceName", serviceName)
			continue
		}

		if strings.HasSuffix(serviceName, ".socket") {
			logger.Info("Warning: service shouldn't be of type socket", "serviceName", serviceName)
			continue
		}

		if strings.HasSuffix(serviceName, ".service") {
			serviceName = strings.TrimSuffix(serviceName, ".service")
		}

		absPath := systemdPath + "/" + serviceName + ".service"

		s[i] = systemdUnit{
			serviceName:  serviceName,
			absPath:      absPath,
			fileContents: unit.File,
		}
	}

	return SystemdUnitConfig{
		units:  s,
		state:  units.State,
		logger: logger,
	}
}

func (s SystemdUnitConfig) Reconcile() error {
	hostFsEnabled := os.Getenv("HOSTFS_ENABLED")
	if hostFsEnabled != "true" {
		err := errors.New("HOSTFS_ENABLED is set to false")
		s.logger.Error(err, "module needs chroot to work, set HOSTFS_ENABLED to true")
		return nil
	}

	if s.state == "present" {
		s.logger.V(1).Info("applying module")
		if err := s.applyConfig(); err != nil {
			return fmt.Errorf("failed to apply module: %w", err)
		}
		s.logger.V(1).Info("module applied")
	} else if s.state == "absent" {
		s.logger.V(1).Info("removing module")
		if err := s.removeModule(); err != nil {
			return fmt.Errorf("failed to remove module: %w", err)
		}
		s.logger.V(1).Info("module removed")
	}

	return nil
}

func (s SystemdUnitConfig) applyConfig() error {
	err := checkOrCreateDirectory(systemdPath)
	if err != nil {
		return fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	isCurrent, err := s.checkCurrentConfig()
	if err != nil {
		return fmt.Errorf("failed to check current configuration: %w", err)
	}

	if isCurrent {
		// do nothing
		return nil
	}

	for _, unit := range s.units {
		err := writeFile(unit.absPath, unit.fileContents)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	// Reload services
	_, err = execChroot("systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("failed to reload daemon: %w", err)
	}

	for _, unit := range s.units {
		_, err := execChroot("systemctl", "start", unit.serviceName)
		if err != nil {
			return fmt.Errorf("failed to start systemd service: %w", err)
		}

		isActive, err := checkIfServiceIsActive(unit.serviceName)
		if err != nil {
			return err
		}

		if !isActive {
			return fmt.Errorf("failed to activate service %s: %w", unit.serviceName, err)
		}
	}
	return nil
}

func (s SystemdUnitConfig) removeModule() error {
	for _, unit := range s.units {
		_, err := execChroot("systemctl", "stop", unit.serviceName)
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			// exit code 5 from "systemd stop service" means
			// that the service is not present in the system
			if ee.ExitCode() != 5 {
				return fmt.Errorf("failed to stop service: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		err = os.Remove(unit.absPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("failed to delete service file: %w", err)
			}
		}
	}

	_, err := execChroot("systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	return nil
}

func (s SystemdUnitConfig) checkCurrentConfig() (bool, error) {
	for _, unit := range s.units {
		isFileEqual, err := checkFileContents(unit.absPath, unit.fileContents)
		if err != nil {
			return false, fmt.Errorf("failed to check file contents: %w", err)
		}

		if !isFileEqual {
			return false, nil
		}

		isActive, err := checkIfServiceIsActive(unit.serviceName)
		if err != nil {
			return false, err
		}

		if !isActive {
			return false, nil
		}
	}

	return true, nil
}
