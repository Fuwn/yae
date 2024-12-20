package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Fuwn/yae/internal/commands"
	"github.com/Fuwn/yae/internal/yae"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

func main() {
	sources := yae.Environment{}

	if err := (&cli.App{
		Name:                 "yae",
		Usage:                "Nix Dependency Manager",
		Description:          "Nix Dependency Manager",
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Fuwn",
				Email: "contact@fuwn.me",
			},
		},
		Before: func(c *cli.Context) error {
			if args := c.Args(); args.Len() == 1 && args.Get(0) == "init" {
				return nil
			}

			location := c.String("sources")

			if _, err := os.Stat(location); os.IsNotExist(err) {
				return fmt.Errorf(
					"file `%s` was not present, run `yae init` to create it",
					location,
				)
			}

			return sources.Load(location)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "sources",
				Value: "./yae.json",
				Usage: "Sources path",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug output",
				Action: func(*cli.Context, bool) error {
					log.SetLevel(log.DebugLevel)

					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "silent",
				Usage: "Silence log output",
				Action: func(*cli.Context, bool) error {
					log.SetLevel(log.WarnLevel)

					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Prevents writing to disk",
			},
		},
		Copyright: fmt.Sprintf("Copyright (c) 2024-%s Fuwn", fmt.Sprint(time.Now().Year())),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				log.Fatal(err.Error())
			}
		},
		Suggest: true,
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Initialise a new Yae environment",
				Action: commands.Init(&sources),
			},
			{
				Name:      "add",
				Args:      true,
				ArgsUsage: "<name> <url>",
				Usage:     "Add a source",
				Flags:     commands.AddFlags(),
				Action:    commands.Add(&sources),
			},
			{
				Name:      "drop",
				ArgsUsage: "<name>",
				Args:      true,
				Usage:     "Drop a source",
				Action:    commands.Drop(&sources),
			},
			{
				Name:      "update",
				Args:      true,
				Usage:     "Update one or all sources",
				ArgsUsage: "[name]",
				Flags:     commands.UpdateFlags(),
				Action:    commands.Update(&sources),
			},
		},
	}).Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
