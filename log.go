package main

import (
	"fmt"
	"io"
	"os"

	"mydocker/path"
)

func logContainer(containerName string) error {
	logFilePath := path.LogPath(containerName)
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
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		return fmt.Errorf("fmt.Fprintf err: %v", err)
	}
	return nil
}
