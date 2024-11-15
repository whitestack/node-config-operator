package modules

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/go-logr/logr"
)

const (
	certPath       = "/host/usr/local/share/ca-certificates/"
	caCertFilePath = "/host/etc/ssl/certs/ca-certificates.crt"
)

// +kubebuilder:object:generate=true
type Certificates struct {
	Certificates []Certificate `json:"certificates"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state"`
}

type Certificate struct {
	FileName string `json:"filename"`
	Content  string `json:"content"`
}

type CertificateConfig struct {
	Certificates
	Log logr.Logger
}

func (c CertificateConfig) Reconcile() error {
	hostFsEnabled := os.Getenv("HOSTFS_ENABLED")
	if hostFsEnabled != "true" {
		err := errors.New("HOSTFS_ENABLED is set to false")
		c.Log.Error(err, "module needs the host's filesystem to work, set HOSTFS_ENABLED to true")
		return nil
	}

	if c.State == "present" {
		c.Log.V(1).Info("applying module")
		if err := c.applyModule(); err != nil {
			return fmt.Errorf("failed to apply module: %w", err)
		}
		c.Log.V(1).Info("module applied")
	} else if c.State == "absent" {
		c.Log.V(1).Info("removing module")
		if err := c.removeModule(); err != nil {
			return fmt.Errorf("failed to remove module: %w", err)
		}
		c.Log.V(1).Info("module removed")
	}

	return nil
}

func (c CertificateConfig) applyModule() error {
	modified := false
	for _, cert := range c.Certificates.Certificates {
		isCurrent, err := checkCurrentConfig(cert.Content)
		if err != nil {
			return fmt.Errorf("failed to check current config: %w", err)
		}

		if isCurrent {
			// do nothing
			continue
		}
		err = writeFile(certPath+cert.FileName, cert.Content)

		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		modified = true
	}
	if modified {
		_, err := execChroot("update-ca-certificates")

		if err != nil {
			return fmt.Errorf("failed to run update-ca-certificates: %w", err)
		}
	}

	return nil
}

func (c CertificateConfig) removeModule() error {
	changed := false

	for _, cert := range c.Certificates.Certificates {
		err := os.Remove(certPath + cert.FileName)

		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("failed to delete file: %w", err)
			} else {
				continue
			}
		}
		changed = true
	}
	if !changed {
		// Skip updating ca certificates if no files were deleted.
		return nil
	}
	_, err := execChroot("update-ca-certificates")
	if err != nil {
		return fmt.Errorf("failed to run update-ca-certificates: %w", err)
	}

	return nil
}

func checkCurrentConfig(content string) (bool, error) {
	doesFileContain, err := checkFileContains(caCertFilePath, content)
	if err != nil {
		return false, err
	}

	if !doesFileContain {
		return false, nil
	}

	return true, nil
}
