package main

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"mydocker/cgroups"
	"mydocker/container"
)

func Run(it bool, resourceConfig *cgroups.ResourceConfig, comArray []string) error {
	parent, writePipe, err := container.NewParentProcessCmd(it)
	if err != nil {
		log.Error("new parent process cmd error:", err)
		return err
	}
	// docker init 成为容器运行的第一个进程
	if err = parent.Start(); err != nil {
		log.Error("parent start error:", err)
		return err
	}
	// 设置资源限制
	err, clearCgroup := enableParentResourceConfig(resourceConfig, parent.Process.Pid)
	if err != nil {
		log.Error("enable parent resource config error:", err)
		return err
	}
	defer clearCgroup()
	// 发送用户命令
	if err = sendUserCommand(comArray, writePipe); err != nil {
		log.Error("send user command error:", err)
		return err
	}
	if err = parent.Wait(); err != nil {
		log.Info("container finished:", err)
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
	userCommamd := strings.Join(comArray, " ")
	if _, err := writePipe.WriteString(userCommamd); err != nil {
		return err
	}
	_ = writePipe.Close()
	return nil
}
