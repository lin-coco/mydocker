package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"mydocker/container"
	"mydocker/path"
)

func stopContainer(f bool, containerName string) error {
	info, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("getContainerPidByName err: %v", err)
	}
	pid, err := strconv.Atoi(info.Pid)
	if err != nil {
		return fmt.Errorf("strconv.Atoi err: %v", err)
	}
	if f {
		// 发送SIGKILL来通知容器停止
		_ = syscall.Kill(pid, syscall.SIGKILL)
	} else {
		// 发送SIGTERM来通知容器停止
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	// 修改容器状态
	info.Status = container.STOP
	info.Pid = " "
	content, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("json.Marshal err: %v", err)
	}
	configFilePath := path.InfoPath(containerName)
	if err = os.WriteFile(configFilePath, content, 0622); err != nil {
		return fmt.Errorf("os.WriteFile err: %v", err)
	}
	return nil
}
