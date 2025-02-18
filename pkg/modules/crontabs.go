package modules

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"os"
)

const (
	crontabsPath = "/host/etc/cron.d"
)

// +kubebuilder:object:generate=true
// Crontabs defines the crontabs section in the NodeConfig resource.
type Crontabs struct {
	Entries []Crontab `json:"entries,omitempty"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state,omitempty"`
}

// Crontab defines an individual crontab entry.
type Crontab struct {
	// Unique identifier for the cron job
	Name string `json:"name"`
	// Special time (reboot, daily, etc.)
	// +kubebuilder:validation:Enum=reboot;yearly;annually;monthly;weekly;daily;hourly
	SpecialTime string `json:"special_time,omitempty"`
	// +default="*"
	// +kubebuilder:default:="*"
	// Minute (default: "*")
	Minute string `json:"minute,omitempty"`
	// +default="*"
	// +kubebuilder:default:="*"
	// Hour (default: "*")
	Hour string `json:"hour,omitempty"`
	// +default="*"
	// +kubebuilder:default:="*"
	// DayOfMonth of the month (default: "*")
	DayOfMonth string `json:"dayOfMonth,omitempty"`
	// +default="*"
	// +kubebuilder:default:="*"
	// Month (default: "*")
	Month string `json:"month,omitempty"`
	// +default="*"
	// +kubebuilder:default:="*"
	// DayOfWeek of the week (default: "*")
	DayOfWeek string `json:"dayOfWeek,omitempty"`
	// Job command or script to execute
	Job string `json:"job"`
	// User under which the task will run
	User string `json:"user"`
}

type CrontabsConfig struct {
	Crontabs
	Log logr.Logger
}

func (c CrontabsConfig) Reconcile() error {
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

func (c CrontabsConfig) applyModule() error {
	// Ensure the cron service is active
	active, err := checkIfServiceIsActive("cron")
	if err != nil {
		return fmt.Errorf("failed to check cron service status: %w", err)
	}
	if !active {
		if err := startCronService(); err != nil {
			return fmt.Errorf("failed to start cron service: %w", err)
		}
	}

	// Apply the cron entries
	for _, entry := range c.Entries {
		c.Log.V(1).Info("Applying crontab entry", "name", entry.Name)
		if err := entry.createCronFile(); err != nil {
			return fmt.Errorf("failed to apply crontab entry '%s': %w", entry.Name, err)
		}
		c.Log.V(1).Info("Crontab applied", "name", entry.Name)
	}
	return nil
}

func (c CrontabsConfig) removeModule() error {
	// Remove the cron entries
	for _, entry := range c.Entries {
		c.Log.V(1).Info("Removing crontab entry", "name", entry.Name)
		if err := entry.removeCronFile(); err != nil {
			return fmt.Errorf("failed to remove crontab entry '%s': %w", entry.Name, err)
		}
		c.Log.V(1).Info("crontab removed", "name", entry.Name)
	}
	return nil
}

func (entry Crontab) createCronFile() error {
	// Sanitize the name to ensure it's a valid filename
	sanitizedName := sanitizeFileName(entry.Name)

	// Build the filename
	fileName := fmt.Sprintf("%s/%s", crontabsPath, sanitizedName)

	// Build the cron line
	var cronLine string
	if entry.SpecialTime != "" {
		cronLine = fmt.Sprintf("@%s %s %s # %s", entry.SpecialTime, entry.User, entry.Job, entry.Name)
	} else {
		cronLine = fmt.Sprintf("%s %s %s %s %s %s %s # %s",
			entry.Minute, entry.Hour, entry.DayOfMonth,
			entry.Month, entry.DayOfWeek, entry.User, entry.Job, entry.Name)
	}

	// Check if the file already exists and has the same content
	contentMatch, err := checkFileContents(fileName, cronLine)
	if err != nil {
		return fmt.Errorf("failed to check file contents for %s: %w", fileName, err)
	}
	if contentMatch {
		return nil // No changes needed
	}

	// Write the file atomically using utils.writeFile
	if err := writeFile(fileName, cronLine); err != nil {
		return fmt.Errorf("failed to write cron file '%s': %w", fileName, err)
	}

	return nil
}

func (entry Crontab) removeCronFile() error {
	// Sanitize the name to ensure it matches the filename
	sanitizedName := sanitizeFileName(entry.Name)
	fileName := fmt.Sprintf("%s/%s", crontabsPath, sanitizedName)

	// Check if the file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil
	}

	// Remove the file
	err := os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("failed to remove cron file '%s': %w", fileName, err)
	}
	return nil
}

func startCronService() error {
	_, err := execChroot("systemctl", "start", "cron")
	if err != nil {
		return fmt.Errorf("failed to start cron service: %w", err)
	}
	isActive, err := checkIfServiceIsActive("cron")
	if err != nil {
		return err
	}

	if !isActive {
		return fmt.Errorf("failed to start cron service: %w", err)
	}
	return nil
}
