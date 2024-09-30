package modules

import (
	"bytes"
	"fmt"

	"github.com/go-logr/logr"
)

// +kubebuilder:object:generate=true
type Hosts struct {
	Hosts []Host `json:"hosts"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state"`
}

type Host struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

type HostModuleConfig struct {
	Hosts
	logger   logr.Logger
	filePath string
}

func NewHostModuleConfig(hosts Hosts, log logr.Logger) HostModuleConfig {
	return HostModuleConfig{
		Hosts:    hosts,
		logger:   log,
		filePath: "/etc/host/hosts",
	}
}

func (c HostModuleConfig) Reconcile() error {
	if c.State == "present" {
		c.logger.V(1).Info("applying module")
		if err := c.applyModule(); err != nil {
			return fmt.Errorf("failed to apply module: %w", err)
		}
		c.logger.V(1).Info("module applied")
	} else if c.State == "absent" {
		c.logger.V(1).Info("removing module")
		if err := c.removeModule(); err != nil {
			return fmt.Errorf("failed to remove module: %w", err)
		}
		c.logger.V(1).Info("module removed")
	}

	return nil
}

func (c HostModuleConfig) applyModule() error {
	blocks := [][]byte{}

	for _, host := range c.Hosts.Hosts {
		blocks = append(blocks, []byte(fmt.Sprintf("%s %s", host.IP, host.Hostname)))
	}

	block := bytes.Join(blocks, []byte("\n"))

	err := writeBlockToFile(c.filePath, []byte{}, []byte{}, block)
	if err != nil {
		return fmt.Errorf("failed to write block to file: %w", err)
	}

	return nil
}

func (c HostModuleConfig) removeModule() error {
	err := deleteBlockFromFile(c.filePath, []byte{}, []byte{})
	if err != nil {
		return fmt.Errorf("failed to delete from file: %w", err)
	}
	return nil
}
