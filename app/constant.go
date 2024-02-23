package app

const (
	Name       = "mydocker"
	Usage      = "mydocker is a simple container runtime implementation.\nThe purpose of this project is to learn how docker works and how to write a docker by ourselves.\nEnjoy it, just for fun."
	UnionPath  = "/root/writedocker/overlay"     // 联合文件系统
	MntPath    = "/root/writedocker/overlay/mnt" // 联合文件系统挂载点；指定容器运行目录
	BusyboxTar = "/root/writedocker/busybox.tar" // busybox tar文件
)
