package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"

	"mydocker/app"
)

func commitContainer(imageName string) error {
	imageTar := path.Join("/root", imageName+".tar")
	cmd := exec.Command("tar", "-czf", imageTar, "-C", app.MntPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run() err: %v", err)
	}
	log.Info(imageTar)
	return nil
}
