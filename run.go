package main

import (
	"fmt"
	"os"
	"strings"

	"mydocker/app"
	"mydocker/cgroups"
	"mydocker/container"
)

func Run(it bool, resourceConfig *cgroups.ResourceConfig, comArray []string) error {
	// parent 父进程启动命令 /proc/self/exe
	parent, writePipe, err := container.NewParentProcessCmd(it)
	if err != nil {
		return fmt.Errorf("container.NewParentProcessCmd err: %v", err)
	}
	// 创建容器的运行空间(文件系统)
	err, clearRunningSpace := container.NewRunningSpace(app.UnionPath, app.MntPath, app.BusyboxTar)
	if err != nil {
		return fmt.Errorf("NewRunningSpace err: %v", err)
	}
	defer clearRunningSpace()
	// 指定运行目录
	parent.Dir = app.MntPath
	// docker init 成为容器运行的第一个进程
	if err = parent.Start(); err != nil {
		return fmt.Errorf("parent.Start() err: %v", err)
	}
	// 设置资源限制
	err, clearCgroup := enableParentResourceConfig(resourceConfig, parent.Process.Pid)
	if err != nil {
		return fmt.Errorf("enableParentResourceConfig err: %v", err)
	}
	defer clearCgroup()
	// 发送用户命令 如 /bin/bash
	if err = sendUserCommand(comArray, writePipe); err != nil {
		return fmt.Errorf("sendUserCommand err: %v", err)
	}
	if err = parent.Wait(); err != nil {
		return fmt.Errorf("parent.Wait() err:%v", err)
	}
	return nil
}

func enableParentResourceConfig(resourceConfig *cgroups.ResourceConfig, parentPid int) (error, func()) {
	cgroupPath, err := cgroups.Create(parentPid)
	clearCgroup := func() {
		_ = cgroups.Clear(cgroupPath)
	}
	defer clearCgroup()
	if err != nil {
		return err, nil
	}
	if err = cgroups.Set(cgroupPath, resourceConfig); err != nil {
		return err, nil
	}
	if err = cgroups.Apply(cgroupPath, parentPid); err != nil {
		return err, nil
	}
	return nil, clearCgroup
}

func sendUserCommand(comArray []string, writePipe *os.File) error {
	userCommand := strings.Join(comArray, " ")
	if _, err := writePipe.WriteString(userCommand); err != nil {
		return err
	}
	_ = writePipe.Close()
	return nil
}
