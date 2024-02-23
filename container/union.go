package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

/*
NewRunningSpace 创建容器运行时文件系统
unionPath: 容器运行联合文件系统
mntPath: Union File System挂载点
busyboxTarPath: tar文件路径
*/
func NewRunningSpace(unionPath string, mntPath string, busyboxTarPath string) (error, func()) {
	clearFunc := func() {
		deleteRunningSpace(unionPath, mntPath)
	}
	lower := path.Join(unionPath, "lowerLayer")
	upper := path.Join(unionPath, "upperLayer")
	worker := path.Join(unionPath, "workerLayer")
	if err := createLowerLayer(lower, busyboxTarPath); err != nil {
		return fmt.Errorf("createLowerLayer err: %v", err), clearFunc
	}
	if err := createUpperLayer(upper); err != nil {
		return fmt.Errorf("createUpperLayer err: %v", err), clearFunc
	}
	if err := createWorkerLayer(worker); err != nil {
		return fmt.Errorf("createWorkerLayer err: %v", err), clearFunc
	}
	if err := createMntPath(mntPath); err != nil {
		return fmt.Errorf("createMntPath err: %v", err), clearFunc
	}
	if err := execMountPoint(lower, upper, worker, mntPath); err != nil {
		return fmt.Errorf("CreateMountPoint err: %v", err), clearFunc
	}
	return nil, clearFunc
}

/*
deleteRunningSpace 删除容器运行时文件系统，退出容器
*/
func deleteRunningSpace(unionPath string, mntPath string) {
	upper := path.Join(unionPath, "upperLayer")
	worker := path.Join(unionPath, "workerLayer")
	if err := deleteMountPoint(mntPath); err != nil {
		log.Errorf("deleteMountPoint err: %v", err)
	}
	if err := deleteMntPath(mntPath); err != nil {
		log.Errorf("deleteMntPath err: %v", err)
	}
	if err := deleteUpperLayer(upper); err != nil {
		log.Errorf("deleteUpperLayer err: %v", err)
	}
	if err := deleteWorkerLayer(worker); err != nil {
		log.Errorf("deleteWorkerLayer err: %v", err)
	}
}

/*
createLowerLayer 创建只读层lower
*/
func createLowerLayer(lowerPath string, busyboxTarPath string) error {
	exist, err := pathExist(lowerPath)
	if err != nil {
		return fmt.Errorf("pathExist err: %v", err)
	}
	if !exist {
		// 解压到busyboxUrl
		if err = os.MkdirAll(lowerPath, 0777); err != nil {
			return fmt.Errorf("os.Mkdir err: %v", err)
		}
		if _, err = exec.Command("tar", "-xvf", busyboxTarPath, "-C", lowerPath).CombinedOutput(); err != nil {
			return fmt.Errorf("exec.Command().CombinedOutput err: %v", err)
		}
	}
	return nil
}

/*
createUpperLayer 创建可写层upper
*/
func createUpperLayer(upperPath string) error {
	if err := os.MkdirAll(upperPath, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir err: %v", err)
	}
	return nil
}

/*
createWorkerLayer 创建工作目录worker
*/
func createWorkerLayer(workerPath string) error {
	if err := os.MkdirAll(workerPath, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir err: %v", err)
	}
	return nil
}

/*
createMntPath 创建mnt挂载目录
*/
func createMntPath(mntPath string) error {
	if err := os.MkdirAll(mntPath, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.MkdirAll err: %v", err)
	}
	return nil
}

/*
execMountPoint 执行挂载
*/
func execMountPoint(lowerPath string, upperPath string, workerPath string, mntPath string) error {
	// 挂载到mnt路径下
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o",
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerPath, upperPath, workerPath), mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run() err: %v", err)
	}
	return nil
}

/*
deleteMntPath 删除挂载目录
*/
func deleteMntPath(mntPath string) error {
	if err := os.RemoveAll(mntPath); err != nil {
		return fmt.Errorf("os.RemoveAll err: %v", err)
	}
	return nil
}

/*
deleteMountPoint 取消挂载mnt
*/
func deleteMountPoint(mntPath string) error {
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run() err: %v", err)
	}
	return nil
}

/*
deleteUpperLayer 删除upper
*/
func deleteUpperLayer(upperPath string) error {
	if err := os.RemoveAll(upperPath); err != nil {
		return fmt.Errorf("os.RemoveAll err: %v", err)
	}
	return nil
}

/*
deleteWorkerLayer 删除worker
*/
func deleteWorkerLayer(workerPath string) error {
	if err := os.RemoveAll(workerPath); err != nil {
		return fmt.Errorf("os.RemoveAll err: %v", err)
	}
	return nil
}

/*
判断路径是否存在
*/
func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
