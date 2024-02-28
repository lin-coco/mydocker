package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"mydocker/container"
	_ "mydocker/nsenter"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func ExecContainer(containerName string, commandArray []string) error {
	// 获取目标容器的pid
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("getContainerPidByName err: %v", err)
	}
	cmdStr := strings.Join(commandArray, " ")
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = os.Setenv(EnvExecPid, pid); err != nil {
		return fmt.Errorf("os.Setenv err: %v", err)
	}
	if err = os.Setenv(EnvExecCmd, cmdStr); err != nil {
		return fmt.Errorf("os.Setenv err: %v", err)
	}
	if err = cmd.Run(); err != nil { // 再次运行docker exec，是为了让C拿到环境变量再执行一次
		return fmt.Errorf("cmd.Run err: %v", err)
	}
	return nil
}

func getContainerPidByName(containerName string) (string, error) {
	configFilePath := filepath.Join(container.DefaultInfoLocation, containerName, container.ConfigName)
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile err: %v", err)
	}
	var info container.Info
	if err = json.Unmarshal(content, &info); err != nil {
		log.Errorf("json.Unmarshal err: %v", err)
	}
	return info.Pid, nil
}
