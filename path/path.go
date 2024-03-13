package path

import (
	"fmt"

	"mydocker/app"
)

const (
	// 镜像容器存储路径
	overlayUnionLocation = "/var/lib/" + app.Name + "/overlay" // 联合文件系统
	imageStoragePath     = overlayUnionLocation + "/image"
	imagePath            = imageStoragePath + "/%s.tar"
	containerUnionPath   = overlayUnionLocation + "/container/%s" // 容器目录（%s为容器名称）
	mntPath              = containerUnionPath + "/mnt"            // 挂载路径 （%s为容器名称）
	lowerPath            = containerUnionPath + "/lower"          // lower路径 （%s为容器名称）
	upperPath            = containerUnionPath + "/upper"          // upper路径 （%s为容器名称）
	workerPath           = containerUnionPath + "/worker"         // worker路径 （%s为容器名称）
	// 容器基本信息
	containerInfoLocation = "/var/run/" + app.Name + "/container"
	containerInfoPath     = containerInfoLocation + "/%s"
	infoPath              = containerInfoPath + "/info.json"
	logPath               = containerInfoPath + "/container.log"
	// 网络配置存储目录
	networkLocation = "/var/run/" + app.Name + "/network"
	networkPath     = networkLocation + "/network"
)

func ImageStoragePath() string {
	return imageStoragePath
}
func ImagePath(imageName string) string {
	return fmt.Sprintf(imagePath, imageName)
}
func ContainerUnionPath(containerName string) string {
	return fmt.Sprintf(containerUnionPath, containerName)
}
func MntPath(containerName string) string {
	return fmt.Sprintf(mntPath, containerName)
}
func LowerPath(containerName string) string {
	return fmt.Sprintf(lowerPath, containerName)
}
func UpperPath(containerName string) string {
	return fmt.Sprintf(upperPath, containerName)
}
func WorkerPath(containerName string) string {
	return fmt.Sprintf(workerPath, containerName)
}
func ContainerInfoLocation() string {
	return containerInfoLocation
}
func ContainerInfoPath(containerName string) string {
	return fmt.Sprintf(containerInfoPath, containerName)
}
func InfoPath(containerName string) string {
	return fmt.Sprintf(infoPath, containerName)
}
func LogPath(containerName string) string {
	return fmt.Sprintf(logPath, containerName)
}

func NetworkPath() string {
	return networkPath
}
