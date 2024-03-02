package main

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"mydocker/cgroups"
	"mydocker/container"
)

var (
	runCommand = cli.Command{
		Name:  "run",
		Usage: "create container with namespace and cgroups limit\nmydocker run -it [command]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "it",
				Usage: "enable tty", // tty指终端
			},
			cli.BoolFlag{
				Name:  "d",
				Usage: "detach container", // 容器脱离
			},
			cli.StringFlag{
				Name:  "v",
				Usage: "volume", // 数据卷
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "container name",
			},
			cli.StringFlag{
				Name:  "m",
				Usage: "memory limit",
			},
			cli.StringFlag{
				Name:  "cpushare",
				Usage: "cpushare limit",
			},
			cli.StringFlag{
				Name:  "cpuset",
				Usage: "cpuset limit",
			},
		},
		/*
			这里是run命令真正执行的函数
			1. 判断参数是否包含command
			2. 获取用户指定的command
			3. 调用Run function去准备启动容器
		*/
		Action: func(ctx *cli.Context) error {
			var err error
			if len(ctx.Args()) < 1 {
				return errors.New("missing container command")
			}
			var comArray []string // 用户命令
			for _, arg := range ctx.Args() {
				comArray = append(comArray, arg)
			}
			it := ctx.Bool("it")
			d := ctx.Bool("d")
			if it && d {
				err = errors.New("it and d parameter can not both provided")
				log.Errorf("docker run err: %v", err)
				return err
			}
			volume := ctx.String("v")
			name := ctx.String("name")
			resourceConfig := &cgroups.ResourceConfig{
				MemoryLimit: ctx.String("m"),
				CpuShare:    ctx.String("cpushare"),
				CpuSet:      ctx.String("cpuset"),
			}
			if err = Run(it, resourceConfig, volume, name, comArray); err != nil {
				log.Error("docker run err:", err)
			}
			return err
		},
	}
	initCommand = cli.Command{
		Name:  "init",
		Usage: "init container process run user's process in container. Do not call it outside",
		/*
			执行容器初始化操作
		*/
		Action: func(ctx *cli.Context) error {
			var err error
			log.Infof("init come on")
			if err := container.RunContainerInitProcess(); err != nil {
				log.Errorf("docker init err: %v", err)
			}
			return err
		},
	}
	commitCommand = cli.Command{
		Name:  "commit",
		Usage: "commit a container into image",
		Action: func(ctx *cli.Context) error {
			var err error
			if len(ctx.Args()) < 2 {
				log.Errorf("missing container name or image name")
				return err
			}
			containerName := ctx.Args().Get(0)
			imageName := ctx.Args().Get(1)
			if err = commitContainer(containerName, imageName); err != nil {
				log.Errorf("docker commit err: %v", err)
			}
			return err
		},
	}
	listCommand = cli.Command{
		Name:  "ps",
		Usage: "list all the containers",
		Action: func(ctx *cli.Context) error {
			var err error
			if err = ListContainers(); err != nil {
				log.Errorf("docker ps err: %v", err)
			}
			return err
		},
	}
	logCommand = cli.Command{
		Name:  "logs",
		Usage: "print logs of a container",
		Action: func(ctx *cli.Context) error {
			var err error
			if len(ctx.Args()) < 1 {
				log.Error("please input your container name")
				return err
			}
			containerName := ctx.Args().Get(0)
			if err = logContainer(containerName); err != nil {
				log.Errorf("docker logs err: %v", err)
			}
			return err
		},
	}
	execCommand = cli.Command{
		Name:  "exec",
		Usage: "exec a command into container",
		Action: func(ctx *cli.Context) error {
			var err error
			// this is a callback
			if os.Getenv(EnvExecPid) != "" {
				log.Infof("pid callback gid %v", os.Getgid())
				return nil
			}
			if len(ctx.Args()) < 2 {
				log.Errorf("missing container name or command")
				return err
			}
			containerName := ctx.Args().Get(0)
			var commandArray []string
			for i := 1; i < len(ctx.Args()); i++ {
				commandArray = append(commandArray, ctx.Args().Get(i))
			}
			if err = ExecContainer(containerName, commandArray); err != nil {
				log.Errorf("docker exec err: %v", err)
			}
			return err
		},
	}
	stopCommand = cli.Command{
		Name:  "stop",
		Usage: "stop a container",
		Action: func(ctx *cli.Context) error {
			var err error
			if len(ctx.Args()) < 1 {
				log.Errorf("missing container name")
				return err
			}
			containerName := ctx.Args().Get(0)
			if err = stopContainer(containerName); err != nil {
				log.Errorf("docker stop err: %v", err)
			}
			return err
		},
	}
	rmCommand = cli.Command{
		Name:  "rm",
		Usage: "rm a container",
		Action: func(ctx *cli.Context) error {
			var err error
			if len(ctx.Args()) < 1 {
				log.Errorf("missing container name")
				return err
			}
			containerName := ctx.Args().Get(0)
			if err = removeContainer(containerName); err != nil {
				log.Errorf("docker stop err: %v", err)
			}
			return err
		},
	}
)
