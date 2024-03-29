package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"mydocker/cgroups"
	"mydocker/container"
	"mydocker/path"
)

func Run(it bool, resourceConfig *cgroups.ResourceConfig, volume string, envs []string, networkName string, portMappings []string, containerName string, imageName string, comArray []string) error {
	var (
		id          = randStringBytes(10)
		volumePaths []string
		pms         [][]string
		err         error
	)
	if containerName == "" { // 用户没有设置名称
		containerName = id
	} else {
		if b, err := isExistContainerName(containerName); err != nil { // 检查容器名称是否重复
			return fmt.Errorf("isExistContainerName err: %v", err)
		} else if b {
			return fmt.Errorf("same container name exists")
		}
	}
	if volume != "" { // 用户需要挂载卷
		volumePaths, err = volumeExtract(volume)
		if err != nil {
			return fmt.Errorf("volumeExtract err: %v", err)
		}
	}
	if len(portMappings) != 0 { // 用户需要端口映射
		for _, p := range portMappings {
			pm := strings.Split(p, ":")
			if len(pm) != 2 {
				return fmt.Errorf("portmapping:%s err", p)
			}
			pms = append(pms, pm)
		}
	}
	// parent 父进程启动命令 /proc/self/exe
	parent, writePipe, err := container.NewParentProcessCmd(it, envs, containerName)
	if err != nil {
		return fmt.Errorf("container.NewParentProcessCmd err: %v", err)
	}
	// 创建容器的运行空间(文件系统)
	err, clearRunningSpace := container.NewRunningSpace(imageName, containerName, volumePaths)
	if err != nil {
		return fmt.Errorf("container.NewRunningSpace err: %v", err)
	}
	// 指定运行目录
	parent.Dir = path.MntPath(containerName)
	// docker init 成为容器运行的第一个进程
	if err = parent.Start(); err != nil {
		return fmt.Errorf("parent.Start err: %v", err)
	}
	// 设置资源限制
	cgroup2Path, err, clearCgroup := enableParentResourceConfig(resourceConfig, parent.Process.Pid)
	if err != nil {
		return fmt.Errorf("enableParentResourceConfig err: %v", err)
	}
	// 记录容器信息
	cInfo, err, clearRecord := recordContainerInfo(id, containerName, parent.Process.Pid, cgroup2Path, volumePaths, networkName, pms, imageName, comArray)
	if err != nil {
		return fmt.Errorf("recordContainerInfo err: %v", err)
	}
	// 连接网络
	if networkName != "" {
		if err = Connect(networkName, cInfo); err != nil {
			return fmt.Errorf("connect err: %v", err)
		}
	}
	// 发送用户命令 如 /bin/bash
	if err = sendUserCommand(comArray, writePipe); err != nil {
		return fmt.Errorf("sendUserCommand err: %v", err)
	}
	if it { // 交互式创建：父进程等待子进程结束
		if err = parent.Wait(); err != nil {
			return fmt.Errorf("parent.Wait err:%v", err)
		}
		clearRecord()
		clearCgroup()
		clearRunningSpace()
		if networkName != "" {
			// 从网络中移除设备
			if err = DisConnect(networkName, cInfo); err != nil {
				return fmt.Errorf("disConnect err: %v", err)
			}
		}
	}
	log.Infof("container running")
	return nil
}

func enableParentResourceConfig(resourceConfig *cgroups.ResourceConfig, parentPid int) (string, error, func()) {
	cgroup2Path, err := cgroups.Create(parentPid)
	if err != nil {
		return "", err, nil
	}
	clearCgroup := func() {
		if err := cgroups.Clear(cgroup2Path); err != nil {
			log.Errorf("cgroups.clear err: %v", err)
		}
	}
	if err = cgroups.Set(cgroup2Path, resourceConfig); err != nil {
		return "", fmt.Errorf("cgroups.set err: %v", err), nil
	}
	if err = cgroups.Apply(cgroup2Path, parentPid); err != nil {
		return "", fmt.Errorf("cgroups.Apply err: %v", err), nil
	}
	return cgroup2Path, nil, clearCgroup
}

func sendUserCommand(comArray []string, writePipe *os.File) error {
	bytes, err := json.Marshal(&comArray)
	if err != nil {
		return fmt.Errorf("json.Marshal err: %v", err)
	}
	if _, err = writePipe.Write(bytes); err != nil {
		return err
	}
	_ = writePipe.Close()
	return nil
}

/*
记录容器信息
*/
func recordContainerInfo(containerId string, containerName string, containerPID int, cgroup2Path string, volumePaths []string, networkName string, portMappings [][]string, imageName string, commandArray []string) (*container.Info, error, func()) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	var command string
	for _, s := range commandArray {
		if len(strings.Split(s, " ")) > 1 {
			command += "\"" + s + "\""
		} else {
			command += s
		}
	}
	info := container.Info{
		Pid:          strconv.Itoa(containerPID),
		Id:           containerId,
		Name:         containerName,
		Command:      command,
		VolumePaths:  volumePaths,
		Cgroup2Path:  cgroup2Path,
		NetworkName:  networkName,
		PortMappings: portMappings,
		ImageName:    imageName,
		CreateTime:   createTime,
		Status:       container.RUNNING,
	}
	infoBytes, err := json.Marshal(&info)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal err: %v", err), nil
	}
	infoDir := path.ContainerInfoPath(containerName)
	if err = os.MkdirAll(infoDir, 0622); err != nil {
		return nil, fmt.Errorf("os.MkdirAll err: %v", err), nil
	}
	clearFunc := func() {
		if err := deleteContainerInfo(infoDir); err != nil {
			log.Errorf("deleteContainerInfo err: %v", err)
		}
	}
	defer func() {
		if err != nil {
			clearFunc()
		}
	}()
	infoPath := path.InfoPath(containerName)
	file, err := os.Create(infoPath)
	if err != nil {
		return nil, fmt.Errorf("os.Create err: %v", err), nil
	}
	defer func() {
		_ = file.Close()
	}()
	if _, err = file.Write(infoBytes); err != nil {
		return nil, fmt.Errorf("file.Write err: %v", err), nil
	}
	return &info, err, clearFunc
}

/*
退出删除容器信息
*/
func deleteContainerInfo(infoDir string) error {
	if err := os.RemoveAll(infoDir); err != nil {
		return fmt.Errorf("os.RemoveAll err: %v", err)
	}
	return nil
}

// 生成容器id
func randStringBytes(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	var s string
	for i := 0; i < n; i++ {
		s += strconv.Itoa(r.Intn(10))
	}
	return s
}

/*
解析volume字符串
*/
func volumeExtract(volume string) ([]string, error) {
	volumeUrls := strings.Split(volume, ":")
	if len(volumeUrls) != 2 || volumeUrls[0] == "" || volumeUrls[1] == "" {
		return nil, errors.New("volume parameter input is not correct")
	}
	return volumeUrls, nil
}

/*
检查容器名称是否重复
*/
func isExistContainerName(name string) (bool, error) {
	containerInfoPath := path.ContainerInfoPath(name)
	_, err := os.Stat(containerInfoPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
