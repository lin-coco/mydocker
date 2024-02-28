package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"mydocker/container"
)

func logContainer(containerName string) error {
	logFilePath := filepath.Join(container.DefaultInfoLocation, containerName, container.ContainerLog)
	file, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("os.Open err: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("io.ReadAll err: %v", err)
	}
	_, err = fmt.Fprintf(os.Stdout, string(content))
	if err != nil {
		return fmt.Errorf("fmt.Fprintf err: %v", err)
	}
	return nil
}
