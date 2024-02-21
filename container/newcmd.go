package container

import (
	"os"
	"os/exec"
	"syscall"
)

/*
NewParentProcessCmd 生成父进程启动命令，也即是容器 /proc/self/exe init [command]
*/
func NewParentProcessCmd(tty bool) (*exec.Cmd, *os.File, error) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	init := exec.Command("/proc/self/exe", "init") // docker init
	// 容器隔离
	init.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET,
	}
	if tty { // 前台运行
		init.Stdin = os.Stdin
		init.Stdout = os.Stdout
		init.Stderr = os.Stderr
	}
	init.ExtraFiles = []*os.File{readPipe}
	init.Dir = "/root/busybox" // 指定运行目录
	return init, writePipe, nil
}
