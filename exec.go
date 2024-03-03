package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"

	"mydocker/container"
	_ "mydocker/nsenter"
	"mydocker/path"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func ExecContainer(containerName string, commandArray []string) error {
	// 获取目标容器的pid
	info, err := getContainerInfoByName(containerName)
	pid := info.Pid
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
	envs, err := getEnvsByPid(pid)
	if err != nil {
		return fmt.Errorf("getEnvsByPid err: %v", err)
	}
	cmd.Env = append(os.Environ(), envs...)
	if err = cmd.Run(); err != nil { // 再次运行docker exec，是为了让C拿到环境变量再执行一次
		return fmt.Errorf("cmd.Run err: %v", err)
	}
	return nil
}

func getContainerInfoByName(containerName string) (*container.Info, error) {
	infoPath := path.InfoPath(containerName)
	content, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile err: %v", err)
	}
	var info container.Info
	if err = json.Unmarshal(content, &info); err != nil {
		log.Errorf("json.Unmarshal err: %v", err)
	}
	return &info, nil
}

func getEnvsByPid(pid string) ([]string, error) {
	// 进程环境变量存放的位置是/proc/{pid}/environ
	filePath := fmt.Sprintf("/proc/%s/environ", pid)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile err: %v", err)
	}
	// 环境变量分隔符是\u0000
	return strings.Split(string(content), "\u0000"), nil
}
