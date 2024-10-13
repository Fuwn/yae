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
	sources := yae.Sources{}

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

			return sources.Load(c.String("sources"))
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
				Name:   "drop",
				Args:   true,
				Usage:  "Drop a source",
				Action: commands.Drop(&sources),
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
