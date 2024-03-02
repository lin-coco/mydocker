package main

import (
	"fmt"
	"os"

	"mydocker/cgroups"
	"mydocker/container"
	"mydocker/path"
)

func removeContainer(containerName string) error {
	info, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("getContainerPidByName err: %v", err)
	}

	// 检查是否是停止容器
	if info.Status != container.STOP {
		return fmt.Errorf("not a stop container")
	}
	// 清理cgroup
	if err = cgroups.Clear(info.Cgroup2Path); err != nil {
		return fmt.Errorf("cgroups.Clear err: %v", err)
	}
	// 删除存储容器信息的路径
	if err = os.RemoveAll(path.ContainerInfoPath(containerName)); err != nil {
		return fmt.Errorf("os.RemoveAll err: %v", err)
	}
	// 删除容器
	container.DeleteRunningSpace(path.ContainerUnionPath(containerName), path.MntPath(containerName), info.VolumePaths)
	return nil
}
