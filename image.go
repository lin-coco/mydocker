package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"mydocker/path"
)

func listImages() error {
	storagePath := path.ImageStoragePath()
	imageNames := make([]string, 0)
	err := filepath.Walk(storagePath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		var name string
		if len(info.Name()) > 4 && info.Name()[len(info.Name())-4:] == ".tar" {
			name = info.Name()[:len(info.Name())-4]
		} else {
			name = info.Name()
		}
		imageNames = append(imageNames, name)
		return nil
	})
	if err != nil {
		return fmt.Errorf("filepath.Walk err: %v", err)
	}
	var res string
	for _, name := range imageNames {
		res += name + " "
	}
	res = res[:len(res)-1]
	// 控制台输出的信息列
	_, err = os.Stdout.Write([]byte(res))
	if err != nil {
		return fmt.Errorf("os.Stdout.Write err: %v", err)
	}
	_ = os.Stdout.Close()
	return nil
}
