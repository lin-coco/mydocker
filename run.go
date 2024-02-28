package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"mydocker/app"
	"mydocker/cgroups"
	"mydocker/container"
)

func Run(it bool, resourceConfig *cgroups.ResourceConfig, volume string, name string, comArray []string) error {
	// parent 父进程启动命令 /proc/self/exe
	parent, writePipe, err := container.NewParentProcessCmd(it)
	if err != nil {
		return fmt.Errorf("container.NewParentProcessCmd err: %v", err)
	}
	// 创建容器的运行空间(文件系统)
	err, clearRunningSpace := container.NewRunningSpace(app.UnionPath, app.MntPath, app.BusyboxTar, volume)
	if err != nil {
		return fmt.Errorf("NewRunningSpace err: %v", err)
	}
	// 指定运行目录
	parent.Dir = app.MntPath
	// docker init 成为容器运行的第一个进程
	if err = parent.Start(); err != nil {
		return fmt.Errorf("parent.Start() err: %v", err)
	}
	// 设置资源限制
	err, clearCgroup := enableParentResourceConfig(resourceConfig, parent.Process.Pid)
	if err != nil {
		return fmt.Errorf("enableParentResourceConfig err: %v", err)
	}
	// 记录容器信息
	err, clearRecord := recordContainerInfo(parent.Process.Pid, name, comArray)
	if err != nil {
		return fmt.Errorf("recordContainerInfo err: %v", err)
	}
	// 发送用户命令 如 /bin/bash
	if err = sendUserCommand(comArray, writePipe); err != nil {
		return fmt.Errorf("sendUserCommand err: %v", err)
	}
	if it { // 交互式创建：父进程等待子进程结束
		if err = parent.Wait(); err != nil {
			return fmt.Errorf("parent.Wait() err:%v", err)
		}
		clearRecord()
		clearCgroup()
		clearRunningSpace()
	}
	return nil
}

func enableParentResourceConfig(resourceConfig *cgroups.ResourceConfig, parentPid int) (error, func()) {
	cgroupPath, err := cgroups.Create(parentPid)
	if err != nil {
		return err, nil
	}
	clearCgroup := func() {
		if err := cgroups.Clear(cgroupPath); err != nil {
			log.Errorf("cgroups.Clear err: %v", err)
		}
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
	userCommand := strings.Join(comArray, " ")
	if _, err := writePipe.WriteString(userCommand); err != nil {
		return err
	}
	_ = writePipe.Close()
	return nil
}

/*
记录容器信息
*/
func recordContainerInfo(containerPID int, containerName string, commandArray []string) (error, func()) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, " ")
	if containerName == "" {
		containerName = id // 用户不指定用户名，就使用id作为用户名
	}
	info := container.Info{
		Pid:        strconv.Itoa(containerPID),
		Id:         id,
		Name:       containerName,
		Command:    command,
		CreateTime: createTime,
		Status:     container.RUNNING,
	}
	infoBytes, err := json.Marshal(&info)
	if err != nil {
		return fmt.Errorf("json.Marshal err: %v", err), nil
	}
	infoDir := path.Join(container.DefaultInfoLocation, containerName)
	if err = os.MkdirAll(infoDir, 0622); err != nil {
		return fmt.Errorf("os.MkdirAll err: %v", err), nil
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
	infoPath := filepath.Join(infoDir, container.ConfigName)
	file, err := os.Create(infoPath)
	if err != nil {
		return fmt.Errorf("os.Create err: %v", err), nil
	}
	defer func() {
		_ = file.Close()
	}()
	if _, err = file.Write(infoBytes); err != nil {
		return fmt.Errorf("file.Write err: %v", err), nil
	}
	return err, clearFunc
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
