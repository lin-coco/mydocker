package cgroups

import (
	"os"
	"path"
	"strconv"
	"syscall"
)

/*
ResourceConfig 用于传递资源限制配置的结构体，包含内存限制，CPU时间片权重，CPU核心数
*/
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

const (
	// 资源配置文件
	memoryLimitFile = "memory.max"
	cpuShareFile    = "cpu.weight"
	cpuSetFile      = "cpuset.cpus"
	// 进程pid配置文件
	cgroupProcsFile = "cgroup.procs"
)

/*
Create 创建cgroup 这里将cgroup抽象成了path，原因是cgroup在hierarchy的路径，便是虚拟文件系统中的虚拟路径
*/
func Create(pid int) (string, error) {
	cgroups2MountPath, err := findCgroups2MountPath()
	if err != nil {
		return "", err
	}
	cgroupPath := getCgroupPath(cgroups2MountPath, pid)
	return cgroupPath, os.Mkdir(cgroupPath, 0755)
}

/*
Set 设置cgroup对于资源的限制
*/
func Set(cgroupPath string, res *ResourceConfig) error {
	if res.MemoryLimit != "" {
		// 对此cgroup设置内存限制
		configFile := path.Join(cgroupPath, memoryLimitFile)
		content := []byte(res.MemoryLimit)
		if err := writeResourceConfigFile(configFile, content); err != nil {
			return err
		}
	}
	if res.CpuShare != "" {
		// 对此cgroup设置cpu时间片权重
		configFile := path.Join(cgroupPath, cpuShareFile)
		content := []byte(res.CpuShare)
		if err := writeResourceConfigFile(configFile, content); err != nil {
			return err
		}
	}
	if res.CpuSet != "" {
		// 对此cgroup设置cpu核心数
		configFile := path.Join(cgroupPath, cpuSetFile)
		content := []byte(res.CpuSet)
		if err := writeResourceConfigFile(configFile, content); err != nil {
			return err
		}
	}
	return nil
}

/*
Clear 删除cgroup便是删除对应的cgroupPath目录
*/
func Clear(cgroupPath string) error {
	return syscall.Rmdir(cgroupPath)
}

/*
Apply 将进程加入到cgroupPath对应的cgroup中
*/
func Apply(cgroupPath string, pid int) error {
	return os.WriteFile(path.Join(cgroupPath, cgroupProcsFile), []byte(strconv.Itoa(pid)), 0644)
}
