package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"mydocker/app"
)

func main() {
	a := cli.NewApp()
	a.Name = app.Name
	a.Usage = app.Usage
	a.Commands = []cli.Command{
		runCommand,
		initCommand,
		commitCommand,
		listCommand,
		logCommand,
		execCommand,
	}
	a.Before = func(context *cli.Context) error {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
