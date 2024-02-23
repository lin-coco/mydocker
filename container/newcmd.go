package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

/*
NewParentProcessCmd 生成父进程启动命令，也即是容器 /proc/self/exe init [command]
*/
func NewParentProcessCmd(it bool, volume string) (*exec.Cmd, *os.File, error) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("os.Pipe err: %v", err)
	}
	init := exec.Command("/proc/self/exe", "init") // docker init
	// 容器隔离
	init.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET,
	}
	if it { // 前台运行
		init.Stdin = os.Stdin
		init.Stdout = os.Stdout
		init.Stderr = os.Stderr
	}
	init.ExtraFiles = []*os.File{readPipe}
	return init, writePipe, nil
}
