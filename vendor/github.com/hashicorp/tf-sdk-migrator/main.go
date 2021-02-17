package main

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/hashicorp/tf-sdk-migrator/cmd/check"
	"github.com/hashicorp/tf-sdk-migrator/cmd/migrate"
	"github.com/hashicorp/tf-sdk-migrator/cmd/v2upgrade"
	"github.com/mitchellh/cli"
)

func main() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	ui := &cli.ColoredUi{
		OutputColor: cli.UiColorBlue,
		InfoColor:   cli.UiColorGreen,
		ErrorColor:  cli.UiColorRed,
		WarnColor:   cli.UiColorYellow,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}

	c := cli.NewCLI("tf-sdk-migrator", "0.1.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		check.CommandName:     check.CommandFactory(ui),
		migrate.CommandName:   migrate.CommandFactory(ui),
		v2upgrade.CommandName: v2upgrade.CommandFactory(ui),
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
