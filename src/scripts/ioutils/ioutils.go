package ioutils

import (
	"errors"
	"os"
)

func AppendToFile(filename, content string) error {
	// Open the file with the O_APPEND and O_WRONLY flags to append data.
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the content to the file.
	_, err = file.WriteString(content)
	return err
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return !info.IsDir()
}

func CreateTempDirWithPermissions(dir, pattern string, perm os.FileMode) (string, error) {
	tempDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(tempDir, perm); err != nil {
		return "", err
	}

	return tempDir, nil
}

func CreateTempFileWithPermissions(dir, pattern string, perm os.FileMode) (*os.File, error) {
	tempFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}

	if err := tempFile.Chmod(perm); err != nil {
		return nil, err
	}

	return tempFile, nil
}
