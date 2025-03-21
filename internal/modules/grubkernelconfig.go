package modules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
)

const (
	grubCfgPath           = "/boot/grub/grub.cfg"
	grubKernelBeginMarker = "# BEGIN MARKER NCO GRUB CONFIG"
	grubKernelEndMarker   = "# END MARKER NCO GRUB CONFIG"
)

// +kubebuilder:object:generate=true
// GrubKernel contains kernel version and command line arguments for GRUB configuration
type GrubKernel struct {
	// KernelVersion specifies the Linux kernel version to be used (e.g. "5.15.0-91-generic")
	KernelVersion string `json:"kernelVersion,omitempty"`
	// CmdlineArgs stores kernel boot parameters to be added to GRUB_CMDLINE_LINUX
	CmdlineArgs []string `json:"args,omitempty"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state,omitempty"`
	// Priority for grub config (default: 50)
	// +kubebuilder:validation:Maximum:=99
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=50
	// +optional
	Priority *int `json:"priority,omitempty"`
}

// IsPresent method checks if the module is present
func (g GrubKernel) IsPresent() bool {
	if g.KernelVersion != "" && len(g.CmdlineArgs) != 0 && g.State == "present" {
		return true
	}
	return false
}

type GrubKernelConfig struct {
	GrubKernel
	Log          logr.Logger
	fileName     string
	prevFileName string
}

func NewGrubKernelConfig(grubKernel GrubKernel, logger logr.Logger, name string) GrubKernelConfig {

	folder := "/host/etc/default/grub.d"
	fileName := fmt.Sprintf("%s/%d-nco-%s.cfg", folder, *grubKernel.Priority, name)
	prevFileName := fmt.Sprintf("%s/99-nco.cfg", folder)

	return GrubKernelConfig{
		GrubKernel:   grubKernel,
		Log:          logger,
		fileName:     fileName,
		prevFileName: prevFileName,
	}
}

// Reconcile applies or removes the GRUB configuration based on the State field.
func (gkc GrubKernelConfig) Reconcile() error {
	hostFsEnabled := os.Getenv("HOSTFS_ENABLED")
	if hostFsEnabled != "true" {
		err := errors.New("HOSTFS_ENABLED is set to false")
		gkc.Log.Error(err, "module needs the host's filesystem to work, set HOSTFS_ENABLED to true")
		return nil
	}

	if gkc.State == "present" {
		gkc.Log.V(1).Info("applying module")
		if err := gkc.applyModule(); err != nil {
			return fmt.Errorf("failed to apply module: %w", err)
		}
		gkc.Log.V(1).Info("module applied")
	} else if gkc.State == "absent" {
		gkc.Log.V(1).Info("removing module")
		if err := gkc.removeModule(); err != nil {
			return fmt.Errorf("failed to remove module: %w", err)
		}
		gkc.Log.V(1).Info("module removed")
	}
	return nil
}

// applyModule applies the GRUB configuration changes.
func (gkc GrubKernelConfig) applyModule() error {
	// remove previous grub config as it's not needed anymore
	if err := deleteFileIfExists(gkc.prevFileName); err != nil {
		return fmt.Errorf("failed to delete prevFileName: %w", err)
	}

	// Build the desired content for the file
	var blockLines []string
	if len(gkc.CmdlineArgs) > 0 {
		cmdlineArgs := strings.Join(gkc.CmdlineArgs, " ")
		blockLines = append(blockLines, fmt.Sprintf("GRUB_CMDLINE_LINUX=\"%s\"", cmdlineArgs))
	}
	if gkc.KernelVersion != "" {
		kernelEntry, err := gkc.findKernelEntry()
		if err != nil {
			return fmt.Errorf("kernel entry not found: %w", err)
		}
		blockLines = append(blockLines, fmt.Sprintf("GRUB_DEFAULT=\"%s\"", kernelEntry))
	}
	desiredBlock := strings.Join(blockLines, "\n")
	if desiredBlock != "" {
		desiredBlock = fmt.Sprintf("%s\n%s\n%s\n", grubKernelBeginMarker, desiredBlock, grubKernelEndMarker)
	}

	// Check if the file already has the desired content
	if desiredBlock != "" {
		matches, err := checkFileContents(gkc.fileName, desiredBlock)
		if err != nil {
			return fmt.Errorf("error checking file contents: %w", err)
		}
		if matches {
			gkc.Log.V(1).Info("GRUB configuration is already in the desired state, no changes needed")
			return nil
		}
	}

	// Verify that the kernel is installed if specified
	if gkc.KernelVersion != "" {
		if err := gkc.ensureKernelInstalled(); err != nil {
			return fmt.Errorf("kernel installation verification failed: %w", err)
		}
	}

	// Write the configuration to the file
	if desiredBlock != "" {
		if err := writeFile(gkc.fileName, desiredBlock); err != nil {
			return fmt.Errorf("error writing GRUB configuration: %w", err)
		}
		gkc.Log.V(1).Info("GRUB configuration updated")
	}

	if err := gkc.runUpdateGrub(); err != nil {
		return fmt.Errorf("error running update-grub: %w", err)
	}
	return nil
}

// removeModule reverts the GRUB configuration changes.
func (gkc GrubKernelConfig) removeModule() error {
	// remove previous grub config as it's not needed anymore
	if err := deleteFileIfExists(gkc.prevFileName); err != nil {
		return fmt.Errorf("failed to delete prevFileName: %w", err)
	}

	// Check if the file exists
	if _, err := os.Stat(gkc.fileName); err == nil {
		// The file exists, delete it
		if err := os.Remove(gkc.fileName); err != nil {
			return fmt.Errorf("error deleting configuration file: %w", err)
		}
		// Run update-grub to apply changes
		if err := gkc.runUpdateGrub(); err != nil {
			return fmt.Errorf("error running update-grub: %w", err)
		}
		gkc.Log.V(1).Info("Configuration file deleted and GRUB updated")
	} else if errors.Is(err, os.ErrNotExist) {
		// The file doesn't exist, no action needed
		gkc.Log.V(1).Info("The file does not exist, no action required")
	} else {
		// Other error when checking the file
		return fmt.Errorf("error checking the file: %w", err)
	}
	return nil
}

// ensureKernelInstalled checks if the specified kernel is installed.
func (gkc GrubKernelConfig) ensureKernelInstalled() error {
	kernelPath := filepath.Join("/host/boot", "vmlinuz-"+gkc.KernelVersion)
	exists, err := checkFileExists(kernelPath)
	if err != nil {
		return fmt.Errorf("error checking kernel installation: %w", err)
	}
	if !exists {
		return fmt.Errorf("kernel version %s is not installed", gkc.KernelVersion)
	}
	return nil
}

// runUpdateGrub runs the update-grub command to apply changes.
func (gkc GrubKernelConfig) runUpdateGrub() error {
	output, err := execChroot("update-grub")
	if err != nil {
		return fmt.Errorf("update-grub failed: %s, output: %s", err, string(output))
	}
	gkc.Log.Info("update-grub executed successfully")
	return nil
}

// findKernelEntry finds the descriptive name of the specified kernel in the GRUB menu.
func (gkc GrubKernelConfig) findKernelEntry() (string, error) {
	// Extract only lines containing the kernel version value
	output, err := execChroot("grep", "menuentry .* "+gkc.KernelVersion, grubCfgPath)
	if err != nil {
		return "", fmt.Errorf("failed to extract menuentry lines from GRUB config: %w", err)
	}
	menuEntry := strings.TrimSpace(string(output))
	if menuEntry == "" {
		return "", fmt.Errorf("kernel entry for version %s not found in GRUB menu", gkc.KernelVersion)
	}
	lines := strings.Split(string(output), "\n")

	// Regular expression to extract the first value within single quotes
	re := regexp.MustCompile(`menuentry '([^']+)'`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip lines containing "recovery mode"
		if strings.Contains(line, "recovery mode") {
			continue
		}
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			entry := match[1]
			return fmt.Sprintf("Advanced options for Ubuntu>%s", entry), nil
		}
	}
	return "", fmt.Errorf("kernel entry for version %s not found in GRUB menu", gkc.KernelVersion)
}
