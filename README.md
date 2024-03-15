# mydocker

## 概述

本项目参考《自己动手写docker》这本书基于Linux系统从0实现的简易的docker

这本书侧重于实战，从实战中解释容器运行的原理，受益匪浅

- [本书github](https://github.com/xianlubird/mydocker)（已给star）
- [得到电子版推荐](https://www.dedao.cn/ebook/detail?id=bODoM61kAj9Rql84gzG5nVNZopXKY3DBKK0JLrBmEDv2QPMOyx7a6e1dbPQj2Zdm)（我是通过得到付费买来电子版看的，微信读书上也有，比得到上贵一点。要找免费的大家应该知道去哪找，不多说。）

## 体验一把

### 拉取代码

```sh
$ git clone https://github.com/lin-coco/mydocker.git
$ cd mydocker
$ go build
```

### 创建一个可访问的nginx容器

我的环境是Linux ubuntu22.04，其他环境没测过

#### 创建mydocker的nginx镜像

nginx镜像创建最快的方式是docker export官方nginx的镜像

```sh
$ docker run -d nginx
$ docker ps
$ docker exec -it 84e655ef17e0 /bin/bash
$ echo "daemon off;" >> /etc/nginx/nginx.conf # 让nginx在容器中前台启动
$ cat /etc/nginx/nginx.conf
$ exit
$ docker export [容器id] -o mynginx.tar
$ mkdir -p /var/lib/mydocker/overlay/image/
$ mv mynginx.tar /var/lib/mydocker/overlay/image/ # 移动到mydocker存储镜像文件的地方
$ ls /var/lib/mydocker/overlay/image/
mynginx.tar
```

#### 创建nginx容器

```sh
$ ./mydocker images
$ sysctl -w net.ipv4.ip_forward=1 # 全局允许forwark转发
$ ./mydocker network create --driver bridge --subnet 192.168.99.0/24 --gateway 192.168.99.1 mynginxbridge # 创建mydocker桥接网络 
$ ./mydocker network list
$ ./mydocker run -d --name bird -net mynginxbridge -p 8888:80 mynginx nginx # 启动nginx镜像
$ ./mydocker ps
$ ps -ef | grep [容器在主机上的pid]
$ ping 192.168.99.2 # 容器ip
$ curl 192.168.99.2:80 # 主机访问容器ip:port
$ curl 10.211.55.9:8888 # 主机ip访问
$ curl 10.211.55.2:8888 # 内网中另外一台机器访问
# 浏览器打开，如下图
```

![image-20240315170822159](https://typora-img-xue.oss-cn-beijing.aliyuncs.com/img/image-20240315170822159.png)

#### 环境还原

```sh
$ ./mydocker stop bird
$ ./mydocker rm bird
$ ./mydocker network remove mynginxbridge
$ sysctl -w "net.ipv4.ip_forward=0"
$ rm -rf /var/lib/mydocker
$ rm -rf /var/run/mydocker
```



### 创建一个flask+redis的计数器

[开发笔记：创建一个flask+redis的计数器](./开发笔记/##创建一个flask+redis的计数器)

## 开发笔记

开发过程中遇到的技术困难总结记录下来

[开发笔记](./开发笔记.md)

- [开发环境](./开发笔记.md/#开发环境)
- [基础技术](./开发笔记.md/#基础技术)
  - [Linux Namespace](./开发笔记.md/##Linux Namespace)
  - [Linux Cgroups](./开发笔记.md/##Linux Cgroups)
  - [Union File System](./开发笔记.md/##Union File System)
- [构造容器](./开发笔记.md/#构造容器)
  - [Linux /proc](./开发笔记.md/##Linux /proc)
  - [logrus](./开发笔记.md/##logrus)
  - [cli](./开发笔记.md/##cli)
  - [mount命令介绍](./开发笔记.md/##mount命令介绍)
  - [cmd/syscall区别](./开发笔记.md/##cmd/syscall区别)
  - [go sys call.Mount](./开发笔记.md/##go sys call.Mount)
  - [go syscall.Exec](./开发笔记.md/##go syscall.Exec)
  - [构造run版本容器](./开发笔记.md/##构造run版本容器)
  - [管道 os.Pipe()](./开发笔记.md/##管道 os.Pipe())
  - [cmd.ExtraFiles](./开发笔记.md/##cmd.ExtraFiles)
  - [增加容器资源限制](./开发笔记.md/##增加容器资源限制)
- [构建镜像](./开发笔记.md/#构建镜像)
  - [busybox](./开发笔记.md/##busybox)
  - [pivot_root](./开发笔记.md/##pivot_root)
  - [syscall.Chdir](./开发笔记.md/##syscall.Chdir)
  - [mount bind(type)](./开发笔记.md/##mount bind(type))
  - [tmpfs](./开发笔记.md/##tmpfs)
  - [使用busybox创建容器](./开发笔记.md/##使用busybox创建容器)
  - [root文件系统](./开发笔记.md/##root文件系统)
  - [使用overlay包装busybox](./开发笔记.md/##实现volume数据卷)
  - [实现volume数据卷](./开发笔记.md/##实现volume数据卷)
  - [实现简单镜像打包](./开发笔记.md/##实现简单镜像打包)
- [构建容器进阶](./开发笔记.md/#构建容器进阶)
  - [tabwriter](./开发笔记.md/##tabwriter)
  - [实现容器的后台运行](./开发笔记.md/##实现容器的后台运行)
  - [孤儿进程与僵尸进程](./开发笔记.md/##孤儿进程与僵尸进程)
  - [实现查看运行中容器](./开发笔记.md/##实现查看运行中容器)
  - [实现查看容器日志](./开发笔记.md/##实现查看容器日志)
  - [setns](./开发笔记.md/##setns)
  - [Cgo](./开发笔记.md/##Cgo)
  - [实现进入容器Namespace](./开发笔记.md/##实现进入容器Namespace)
  - [实现停止容器](./开发笔记.md/##实现停止容器)
  - [docker 镜像和容器的存储路径](./开发笔记.md/##docker 镜像和容器的存储路径)
  - [实现删除容器](./开发笔记.md/##实现删除容器)
  - [实现通过容器制造镜像](./开发笔记.md/##实现通过容器制造镜像)
  - [实现容器指定环境变量运行](./开发笔记.md/##实现容器指定环境变量运行)
  - [网络虚拟化技术](./开发笔记.md/##网络虚拟化技术)
- [容器网络](./开发笔记.md/#容器网络)
  - [ip 命令总结](./开发笔记.md/##ip 命令总结)
  - [Linux iptables](./开发笔记.md/##Linux iptables)
  - [iptables 命令总结](./开发笔记.md/##iptables 命令总结)
  - [Linux Veth](./开发笔记.md/##Linux Veth)
  - [Linux Bridge](./开发笔记.md/##Linux Bridge)
  - [构建docker网络模型](./开发笔记.md/##构建docker网络模型)
  - [net库](./开发笔记.md/##net库)
  - [github.com/vishvananda/netlink库](./开发笔记.md/##github.com/vishvananda/netlink库)
  - [github.com/vishvananda/netns库](./开发笔记.md/##github.com/vishvananda/netns库)
  - [实现docker网络模型](./开发笔记.md/##实现docker网络模型)
- [高级实践](./开发笔记.md/#高级实践)
  - [创建一个可访问的nginx容器](./开发笔记.md/##创建一个可访问的nginx容器)
  - [创建一个flask+redis的计数器](./开发笔记.md/##创建一个flask+redis的计数器)



## 预计多久完成？

个人2024年1月开始，结束3月。累计，1559分钟，约26小时。

得到这个平台统计的时间是没有啥参考性的，因为这是本实战类书籍，还有很多时间在编写代码调试bug，实际时常应该更多。

其实这两个月也不是都在实战这本书，所以拖到两个月时间，个人感觉自己有更多时间的话1个月以内是肯定能读完写完的。



![image-20240315135859548](https://typora-img-xue.oss-cn-beijing.aliyuncs.com/img/image-20240315135859548.png)