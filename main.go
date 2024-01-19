package main

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"mydocker/container"
)

const (
	usage = "mydocker is a simple container runtime implementation.\nThe purpose of this project is to learn how docker works and how to write a docker by ourselves.\nEnjoy it, just for fun."
)

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage
	app.Commands = []cli.Command{
		runCommand,
		initCommand,
	}
	app.Before = func(context *cli.Context) error {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var (
	runCommand = cli.Command{
		Name:  "run",
		Usage: "create container with namespace and cgroups limit\nmydocker run -it [command]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "it",
				Usage: "enable tty",
			},
		},
		/*
			这里是run命令真正执行的函数
			1. 判断参数是否包含command
			2. 获取用户指定的command
			3. 调用Run function去准备启动容器
		*/
		Action: func(ctx *cli.Context) error {
			if len(ctx.Args()) < 1 {
				return errors.New("missing container command")
			}
			cmd := ctx.Args().Get(0)
			tty := ctx.Bool("it") // it 表示交互式终端
			Run(tty, cmd)
			return nil
		},
	}
	initCommand = cli.Command{
		Name:  "init",
		Usage: "init container process run user's process in container. Do not call it outside",
		/*
			1. 获取传递过来的command参数
			2.执行容器初始化操作
		*/
		Action: func(ctx *cli.Context) error {
			log.Infof("init come on")
			cmd := ctx.Args().Get(0)
			log.Infof("command %s", cmd)
			err := container.RunContainerInitProcess(cmd, nil)
			return err
		},
	}
)
