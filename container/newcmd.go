package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"mydocker/path"
)

type Info struct {
	Pid          string     `json:"pid,omitempty"`        // 容器在宿主机上的Pid
	Id           string     `json:"id,omitempty"`         // 容器id
	Name         string     `json:"name,omitempty"`       // 容器名
	Command      string     `json:"command,omitempty"`    // 容器内init进程的运行命令
	VolumePaths  []string   `json:"volumePaths"`          // 挂载的数据卷
	Cgroup2Path  string     `json:"cgroup2Path"`          // cgroup路径
	ImageName    string     `json:"imageName"`            // image名称
	NetworkName  string     `json:"networkName"`          // 网络名称
	PortMappings [][]string `json:"portMappings"`         // 端口映射
	CreateTime   string     `json:"createTime,omitempty"` // 创建时间
	Status       string     `json:"status,omitempty"`     // 容器状态
}

/*
NewParentProcessCmd 生成父进程启动命令，也即是容器 /proc/self/exe init [command]
*/
func NewParentProcessCmd(it bool, envs []string, containerName string) (*exec.Cmd, *os.File, error) {
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
		if err := os.MkdirAll(path.ContainerInfoPath(containerName), 0622); err != nil {
			return nil, nil, fmt.Errorf("os.MkdirAll err: %v", err)
		}
		logPath := path.LogPath(containerName)
		file, err := os.Create(logPath)
		if err != nil {
			return nil, nil, fmt.Errorf("os.Create err: %v", err)
		}
		// 容器输出定向到container log文件
		init.Stdout = file
	}
	init.ExtraFiles = []*os.File{readPipe}
	init.Env = append(os.Environ(), envs...) // 设置进程的环境变量
	return init, writePipe, nil
}
