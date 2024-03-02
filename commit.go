package main

import (
	"fmt"
	"os"
	"os/exec"
	path2 "path"

	"mydocker/path"

	log "github.com/sirupsen/logrus"
)

func commitContainer(containerName string, imageName string) error {
	imageTar := path2.Join("/root", imageName+".tar")
	cmd := exec.Command("tar", "-czf", imageTar, "-C", path.MntPath(containerName), ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run() err: %v", err)
	}
	log.Info(imageTar)
	return nil
}
