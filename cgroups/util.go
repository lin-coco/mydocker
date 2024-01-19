package cgroups

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"mydocker/app"
)

/*
将资源配置写入文件
*/
func writeResourceConfigFile(configFile string, content []byte) error {
	return os.WriteFile(configFile, content, 0644)
}

/*
查找Cgroup2挂载路径
*/
func findCgroups2MountPath() (string, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		for _, field := range fields {
			if field == "cgroup2" {
				return path.Join(fields[4], "system.slice"), nil
			}
		}
	}
	return "", errors.New("no cgroup2 mount found")
}

/*
获取目标cgroupPath
*/
func getCgroupPath(cgroupMountPath string, pid int) string {
	return path.Join(cgroupMountPath, fmt.Sprintf("%s-%s-%d", app.Name, time.Now().Format("20060102150405"), pid))
}
