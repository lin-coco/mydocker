package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"mydocker/container"
)

func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	// docker init 成为容器运行的第一个进程
	if err := parent.Start(); err != nil {
		log.Error("parent start error:", err)
	}
	_ = parent.Wait()
	os.Exit(-1)
}
