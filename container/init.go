package container

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
RunContainerInitProcess
容器进程初始化 使用mount先去挂载proc文件系统，以便后面通过ps等系统命令去查看当前进程资源的情况。
替换容器当前进程 syscall.Exec成为当前pid为1的进程
*/
func RunContainerInitProcess() error {
	userCommand, err := readUserCommand()
	if err != nil {
		log.Error("read user command error:", err)
		return err
	}
	if len(userCommand) == 0 {
		return errors.New("read userCommand is nil")
	}
	// MS_NOEXEC 在本文件系统中不允许运行其他程序。
	// MS_NOSUID 在本系统中运行程序的时候，不允许set-user-ID或set-group-ID。
	// MS_NODEV 这个参数是自从Linux 2.4以来，所有 mount 的系统都会默认设定的参数。
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// 调用exec.LookPath，可以在系统的PATH里面寻找命令的绝对路径
	path, err := exec.LookPath(userCommand[0])
	if err != nil {
		log.Error("exec.LookPath error:", err.Error())
		return err
	}
	if err := syscall.Exec(path, userCommand, os.Environ()); err != nil {
		log.Error("syscall.Exec error:", err.Error())
		return err
	}
	return nil
}

func readUserCommand() ([]string, error) {
	pipe := os.NewFile(uintptr(3), "pipe")
	defer func() {
		_ = pipe.Close()
	}()
	data := make([]byte, 200)
	n, err := pipe.Read(data)
	data = data[:n]
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), " "), nil
}
