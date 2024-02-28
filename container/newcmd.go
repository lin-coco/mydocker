package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"mydocker/app"
)

var (
	RUNNING string = "running"
	STOP    string = "stop"
	Exit    string = "exited"
)
var (
	DefaultInfoLocation string = "/var/run/" + app.Name
	ConfigName          string = "config.json"
	ContainerLog        string = "container.log"
)

type Info struct {
	Pid        string `json:"pid,omitempty"`        // 容器在宿主机上的Pid
	Id         string `json:"id,omitempty"`         // 容器id
	Name       string `json:"name,omitempty"`       // 容器名
	Command    string `json:"command,omitempty"`    // 容器内init进程的运行命令
	CreateTime string `json:"createTime,omitempty"` // 创建时间
	Status     string `json:"status,omitempty"`     // 容器状态
}

/*
NewParentProcessCmd 生成父进程启动命令，也即是容器 /proc/self/exe init [command]
*/
func NewParentProcessCmd(it bool, containerName string) (*exec.Cmd, *os.File, error) {
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
	} else { // 后台运行
		// 生成容器对应的log文件
		if err := os.MkdirAll(path.Join(DefaultInfoLocation, containerName), 0622); err != nil {
			return nil, nil, fmt.Errorf("os.MkdirAll err: %v", err)
		}
		logPath := filepath.Join(DefaultInfoLocation, containerName, ContainerLog)
		file, err := os.Create(logPath)
		if err != nil {
			return nil, nil, fmt.Errorf("os.Create err: %v", err)
		}
		// 容器输出定向到container log文件
		init.Stdout = file
	}
	init.ExtraFiles = []*os.File{readPipe}
	return init, writePipe, nil
}
