package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"mydocker/container"
)

func ListContainers() error {
	// 读取容器存储目录下的所有文件
	entries, err := os.ReadDir(container.DefaultInfoLocation)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.ReadDir err: %v", err)
	}
	infos := make([]container.Info, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			containerName := entry.Name()
			configFilePath := filepath.Join(container.DefaultInfoLocation, containerName, container.ConfigName)
			content, err := os.ReadFile(configFilePath)
			if err != nil {
				return fmt.Errorf("os.ReadFile err: %v", err)
			}
			var info container.Info
			if err = json.Unmarshal(content, &info); err != nil {
				return fmt.Errorf("json.Unmarshal err: %v", err)
			}
			infos = append(infos, info)
		}
	}
	writer := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	// 控制台输出的信息列
	_, err = fmt.Fprintf(writer, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		return fmt.Errorf("fmt.Fprintf: %v", err)
	}
	for _, info := range infos {
		_, err = fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", info.Id, info.Name, info.Pid, info.Status, info.Command, info.CreateTime)
		if err != nil {
			return fmt.Errorf("fmt.Fprintf: %v", err)
		}
	}
	// 刷新，使容器列表打印出来
	if err = writer.Flush(); err != nil {
		return fmt.Errorf("flush err: %v", err)
	}
	return nil
}
