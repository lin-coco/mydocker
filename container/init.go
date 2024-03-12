package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
RunContainerInitProcess
容器进程初始化 使用mount先去挂载proc文件系统，以便后面通过ps等系统命令去查看当前进程资源的情况。
替换容器当前进程 syscall.Exec成为当前pid为1的进程
*/
func RunContainerInitProcess() error {
	// 从管道中获取用户命令
	userCommand, err := readUserCommand()
	if err != nil {
		return fmt.Errorf("readUserCommand err: %v", err)
	}
	if len(userCommand) == 0 {
		return errors.New("len(userCommand) = 0")
	}
	// 给容器做一些挂载
	if err = setUpMount(); err != nil {
		return fmt.Errorf("setUpMount err: %v", err)
	}
	// 执行用户命令
	cmdPath, err := exec.LookPath(userCommand[0]) // 调用exec.LookPath，可以在系统的PATH里面寻找命令的绝对路径
	if err != nil {
		return fmt.Errorf("exec.LookPath err: %v", err)
	}
	if err = syscall.Exec(cmdPath, userCommand, os.Environ()); err != nil {
		return fmt.Errorf("syscall.Exec err: %v", err)
	}
	return nil
}

/*
容器初始化 挂载点
*/
func setUpMount() error {
	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd err: %v", err)
	}
	log.Infof("current location is %s", pwd)
	if err = pivotRoot(pwd); err != nil {
		return fmt.Errorf("pivotRoot err: %v", err)
	}
	// mount proc
	if err = syscall.Mount("proc", "/proc", "proc", syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, ""); err != nil {
		return fmt.Errorf("syscall.Mount err: %v", err)
	}
	// mount tmpfs
	if err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf("syscall.Mount err: %v", err)
	}
	return nil
}

/*
pivotRoot 切换根文件系统
*/
func pivotRoot(root string) error {
	var err error
	/*
		systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示
		声明你要这个新的mount namespace独立。
	*/
	if err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("syscall.Mount err: %v", err)
	}
	/*
		为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount了一次，
		bind mount是把相同的内容换了一个挂载点的挂载方法。
	*/
	//if err = syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
	//	return fmt.Errorf("syscall.Mount root err: %v", err)
	//}
	// 创建rootfs/.pivot_root存储old_root
	pivotDir := path.Join(root, ".pivot_root")
	if err = os.Mkdir(pivotDir, 0777); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.Mkdir err: %v", err)
	}
	if err = syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("syscall.PivotRoot err: %v", err)
	}
	// 修改当前的工作目录到根目录
	if err = syscall.Chdir("/"); err != nil {
		return fmt.Errorf("syscall.Chdir err: %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	//umount rootfs/.pivot_root
	if err = syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("syscall.Unmount err: %v", err)
	}
	// 删除临时文件夹
	if err = os.Remove(pivotDir); err != nil {
		return fmt.Errorf("os.Remove err: %v", err)
	}
	return nil
}

func readUserCommand() ([]string, error) {
	pipe := os.NewFile(uintptr(3), "pipe")
	defer func() {
		_ = pipe.Close()
	}()
	data := make([]byte, 200)
	n, err := pipe.Read(data)
	data = data[:n]
	if err != nil {
		return nil, fmt.Errorf("readUserCommand err: %v", err)
	}
	userCommand := make([]string, 0)
	if err = json.Unmarshal(data, &userCommand); err != nil {
		return nil, fmt.Errorf("json.Unmarshal err: %v", err)
	}
	return userCommand, nil
}
