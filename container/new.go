package container

import (
	"os"
	"os/exec"
	"syscall"
)

/*
NewParentProcess 生成父进程启动命令，也即是容器 /proc/self/exe init [command]
*/
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...) // docker init ...
	// 容器隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
