package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"

	"mydocker/cgroups"
	"mydocker/container"
	"mydocker/path"
)

func removeContainer(f bool, containerName string) error {
	info, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("getContainerPidByName err: %v", err)
	}

	// 检查是否是停止容器
	if !f {
		if info.Status != container.STOP {
			return fmt.Errorf("not a stop container")
		}
	} else {
		var pid int
		pid, err = strconv.Atoi(info.Pid)
		if err != nil {
			return fmt.Errorf("strconv.Atoi err: %v", err)
		}
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	// 清理cgroup
	if err = cgroups.Clear(info.Cgroup2Path); err != nil {
		return fmt.Errorf("cgroups.clear err: %v", err)
	}
	// 删除存储容器信息的路径
	if err = os.RemoveAll(path.ContainerInfoPath(containerName)); err != nil {
		// 先执行umount命令
		cmd := exec.Command("umount", path.MntPath(containerName))
		if e := cmd.Run(); e != nil { // umount 运行成功
			return fmt.Errorf("os.RemoveAll err: %v", err)
		} else {
			log.Infof("exec umount %s", path.MntPath(containerName))
			// 再次删除
			if err = os.RemoveAll(path.ContainerInfoPath(containerName)); err != nil {
				return fmt.Errorf("os.RemoveAll err: %v", err)
			}
		}
	}
	// 删除网络
	if err = DisConnect(info.NetworkName, info); err != nil {
		return fmt.Errorf("DisConnect err: %v", err)
	}
	// 删除容器
	container.DeleteRunningSpace(path.ContainerUnionPath(containerName), path.MntPath(containerName), info.VolumePaths)
	return nil
}
