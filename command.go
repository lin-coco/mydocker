package main

import (
	"errors"

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
			cli.StringFlag{
				Name:  "v",
				Usage: "volume", // 数据卷
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
			volume := ctx.String("v")
			resourceConfig := &cgroups.ResourceConfig{
				MemoryLimit: ctx.String("m"),
				CpuShare:    ctx.String("cpushare"),
				CpuSet:      ctx.String("cpuset"),
			}
			if err = Run(it, resourceConfig, comArray, volume); err != nil {
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
			if len(ctx.Args()) < 1 {
				log.Errorf("Missing container name")
			}
			imageName := ctx.Args().Get(0)
			if err = commitContainer(imageName); err != nil {
				log.Errorf("commitContainer err: %v", err)
			}
			return err
		},
	}
)
