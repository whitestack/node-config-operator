package modules

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

// +kubebuilder:object:generate=true
type BlockInFiles struct {
	Blocks []BlockInFile `json:"blocks"`
	// +kubebuilder:Enum="present";"absent"
	State string `json:"state"`
}

type BlockInFile struct {
	FileName string `json:"filename"`
	Content  string `json:"content"`
	// +default="# BEGIN MARKER NCO"
	// +kubebuilder:default:="# BEGIN MARKER NCO"
	// Marker that signals the start of a block
	BeginMarker string `json:"beginMarker"`
	// +default="# END MARKER NCO"
	// +kubebuilder:default:="# END MARKER NCO"
	// Marker that signals the end of the block
	EndMarker string `json:"endMarker"`
}

type BlockInFileConfig struct {
	BlockInFiles
	Log logr.Logger
}

func (b BlockInFileConfig) Reconcile() error {
	hostFsEnabled := os.Getenv("HOSTFS_ENABLED")
	if hostFsEnabled != "true" {
		err := errors.New("HOSTFS_ENABLED is set to false")
		b.Log.Error(err, "module needs the host's filesystem to work, set HOSTFS_ENABLED to true")
		return nil
	}

	if b.State == "present" {
		b.Log.V(1).Info("applying module")
		if err := b.applyModule(); err != nil {
			return fmt.Errorf("failed to apply module: %w", err)
		}
		b.Log.V(1).Info("module applied")
	} else if b.State == "absent" {
		b.Log.V(1).Info("removing module")
		if err := b.removeModule(); err != nil {
			return fmt.Errorf("failed to remove module: %w", err)
		}
		b.Log.V(1).Info("module removed")
	}

	return nil
}

func (b BlockInFileConfig) applyModule() error {
	for _, block := range b.Blocks {
		err := writeBlockToFile(
			// Write file to host's filesystem
			"/host"+block.FileName,
			[]byte(block.BeginMarker),
			[]byte(block.EndMarker),
			[]byte(block.Content),
		)

		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}

func (b BlockInFileConfig) removeModule() error {
	for _, block := range b.Blocks {
		err := deleteBlockFromFile(
			"/host"+block.FileName,
			[]byte(block.BeginMarker),
			[]byte(block.EndMarker),
		)

		if err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
	}

	return nil
}
