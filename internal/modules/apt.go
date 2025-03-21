package modules

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/go-logr/logr"
)

// +kubebuilder:object:generate=true
type AptPackages struct {
	Packages []AptPackage `json:"packages,omitempty"`
	// +kubebuilder:validation:Enum="present";"absent"
	State string `json:"state,omitempty"`
}

// IsPresent method checks if the module is present
func (a AptPackages) IsPresent() bool {
	if len(a.Packages) != 0 && a.State == "present" {
		return true
	}
	return false
}

type AptPackage struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type AptModuleConfig struct {
	AptPackages
	Logger logr.Logger
}

func (a AptModuleConfig) Reconcile() error {
	hostFsEnabled := os.Getenv("HOSTFS_ENABLED")
	if hostFsEnabled != "true" {
		err := errors.New("HOSTFS_ENABLED is set to false")
		a.Logger.Error(err, "module needs chroot to work, set HOSTFS_ENABLED to true")
		return nil
	}

	aptEnabled := os.Getenv("APT_ENABLED")
	if aptEnabled != "true" {
		err := errors.New("APT_ENABLED is set to false")
		a.Logger.Error(err, "set APT_ENABLED to true to enable apt module")
		return nil
	}

	moduleError := ModuleError{"aptPackages", nil}
	if a.State == "present" {
		a.Logger.V(1).Info("applying module")
		if err := a.applyModule(); err != nil {
			moduleError.error = err
			return moduleError
		}
		a.Logger.V(1).Info("module applied")
	} else if a.State == "absent" {
		a.Logger.V(1).Info("removing module")
		if err := a.removeModule(); err != nil {
			moduleError.error = err
			return moduleError
		}
		a.Logger.V(1).Info("module removed")
	}

	return nil
}

func (a AptModuleConfig) applyModule() error {
	installCmd := []string{"apt-get", "install", "-y", "--allow-downgrades"}
	for _, pkg := range a.Packages {
		pkgName := pkg.Name
		if pkg.Version != "" {
			pkgName = pkgName + "=" + pkg.Version
		}
		installCmd = append(installCmd, pkgName)
	}

	output, err := execChroot(installCmd...)
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			aptErrors, err := getAptErrors(output)
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("apt errors: %s", bytes.Join(aptErrors, []byte{' '}))
			return errors.New(msg)
		}
		return err
	}

	return nil
}

func (a AptModuleConfig) removeModule() error {
	a.Logger.Info("nothing to do")
	return nil
}

func getAptErrors(input []byte) ([][]byte, error) {
	r := regexp.MustCompile("(?m)^E:.*")
	output := r.FindAll(input, -1)
	if output == nil {
		return nil, errors.New("no error found in apt stderr")
	}

	return output, nil
}

func AptUpdate() error {
	output, err := execChroot("apt-get", "update", "-y")
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			aptErrors, err := getAptErrors(output)
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("apt errors: %s", bytes.Join(aptErrors, []byte{' '}))
			return errors.New(msg)
		}
		return err
	}

	return nil
}
