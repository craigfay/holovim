package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func getExecDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	return filepath.Dir(execPath), nil
}

func getClosestDir(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", path)
		}
		return "", fmt.Errorf("error accessing path: %w", err)
	}

	// If the path is a directory, return it as-is
	if info.IsDir() {
		return path, nil
	}

	// Otherwise, return the parent directory
	return filepath.Dir(path), nil
}

type FileStatus int

const (
	FileStatusNotExists FileStatus = iota
	FileStatusAccessDenied
	FileStatusIsDirectory
	FileStatusIsFile
)

func checkPath(path string) FileStatus {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return FileStatusNotExists
	}

	if err != nil {
		return FileStatusAccessDenied
	}

	if info.IsDir() {
		return FileStatusIsDirectory
	} else {
		return FileStatusIsFile
	}
}
