package modules

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

// Interface that all modules implement
type Config interface {
	Reconcile() error
}

func writeFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func checkFileContents(filePath, lines string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	if strings.Trim(string(content), "\n") != strings.Trim(lines, "\n") {
		return false, nil
	}

	return true, nil
}

func checkOrCreateDirectory(path string) error {
	_, err := os.Stat(path)
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return nil
}

func writeBlockToFile(path string, beginMarker, endMarker []byte, block []byte) error {
	// Default values
	if len(beginMarker) == 0 {
		beginMarker = []byte("# BEGIN MARKER NCO")
	}
	if len(endMarker) == 0 {
		endMarker = []byte("# END MARKER NCO")
	}

	fileRead, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer fileRead.Close()

	lines, err := writeBlock(fileRead, beginMarker, endMarker, block)
	if err != nil {
		return fmt.Errorf("failed to write block: %w", err)
	}

	fileWrite, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error opening file for write: %w", err)
	}

	_, err = fileWrite.Write(lines)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	defer fileWrite.Close()

	return nil
}

func writeBlock(reader io.Reader, beginMarker, endMarker, block []byte) ([]byte, error) {
	newLines := [][]byte{}
	found := false
	insideBlock := false

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !found && bytes.Equal(line, beginMarker) {
			found = true
			insideBlock = true

			newLines = append(newLines, beginMarker, block)

			continue
		} else if insideBlock && bytes.Equal(line, endMarker) {
			insideBlock = false

			newLines = append(newLines, endMarker)

			continue
		}

		if !insideBlock {
			newLines = append(newLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if found && insideBlock {
		// Begin marker found but no end marker found, add
		// end marker to end of file
		newLines = append(newLines, endMarker)
	} else if !found {
		// Add block with markers at the end of file
		newLines = append(newLines, beginMarker, block, endMarker)
	}

	out := bytes.Join(newLines, []byte("\n"))
	return out, nil
}

func deleteBlockFromFile(path string, beginMarker, endMarker []byte) error {
	// Default values
	if len(beginMarker) == 0 {
		beginMarker = []byte("# BEGIN MARKER NCO")
	}
	if len(endMarker) == 0 {
		endMarker = []byte("# END MARKER NCO")
	}

	fileRead, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer fileRead.Close()

	lines, err := deleteBlock(fileRead, beginMarker, endMarker)

	fileWrite, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error opening file for write: %w", err)
	}

	_, err = fileWrite.Write(lines)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	defer fileWrite.Close()

	return nil
}

func deleteBlock(reader io.Reader, beginMarker, endMarker []byte) ([]byte, error) {
	newLines := [][]byte{}
	insideBlock := false

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !insideBlock && bytes.Equal(line, beginMarker) {
			insideBlock = true
			continue
		} else if insideBlock && bytes.Equal(line, endMarker) {
			insideBlock = false
			continue
		}

		if !insideBlock {
			newLines = append(newLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	out := bytes.Join(newLines, []byte("\n"))
	return out, nil
}

func execChroot(args ...string) ([]byte, error) {
	cmdArgs := append([]string{"/host"}, args...)
	cmd := exec.Command("chroot", cmdArgs...)
	return cmd.CombinedOutput()
}
