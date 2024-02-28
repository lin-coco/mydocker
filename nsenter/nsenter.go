package nsenter

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
	// 从环境变量中获取需要进入容器的pid
	fprintf(stdout, "enter_namespace start!");
	char *mydocker_pid;
	mydocker_pid = getenv("mydocker_pid");
	if (!mydocker_pid) {
		fprintf(stdout, "missing mydocker_pid env skip nsenter");
		return;
	}
	// 从环境变量中获取需要执行的命令
	char *mydocker_cmd;
	mydocker_cmd = getenv("mydocker_cmd");
	if (!mydocker_cmd) {
		fprintf(stdout, "missing mydocker_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	// 需要进入的五种Namespace
	char *namespace[] = {"ipc","uts","net","pid","mnt"};
	for (i=0;i < 5;i++) {
		// 拼接对应的路径
		sprintf(nspath,"/proc/%s/ns/%s",mydocker_pid,namespace[i]);
		int fd = open(nspath,O_RDONLY);
		// 当一个进程调用 setns() 并成功设置其命名空间后，该进程及其后续通过 fork() 创建的子进程都将位于这个新的命名空间中
		if (setns(fd,0) == -1) {
			fprintf(stderr, "setns on %s namespace failed: %s\n", namespace[i], strerror(errno));
		}
		close(fd);
	}
	// 执行指令
	int res = system(mydocker_cmd);
	exit(0);
}
*/
import "C" // 必须紧贴在代码下面
