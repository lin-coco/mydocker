package container

import (
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
RunContainerInitProcess
容器进程初始化 使用mount先去挂载proc文件系统，以便后面通过ps等系统命令去查看当前进程资源的情况。
替换容器当前进程 syscall.Exec成为当前pid为1的进程
*/
func RunContainerInitProcess(command string, args []string) error {
	log.Infof("command %s", command)
	// MS_NOEXEC 在本文件系统中不允许运行其他程序。
	// MS_NOSUID 在本系统中运行程序的时候，不允许set-user-ID或set-group-ID。
	// MS_NODEV 这个参数是自从Linux 2.4以来，所有 mount 的系统都会默认设定的参数。
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
