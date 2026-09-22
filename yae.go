package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fuwn/yae/internal/commands"
	"github.com/Fuwn/yae/internal/yae"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

var Version string

func main() {
	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	if err := newApp().RunContext(context, os.Args); err != nil {
		log.Fatal(err.Error())
	}
}

func newApp() *cli.App {
	sources := yae.Environment{}
	loadSources := func(context *cli.Context) error {
		if err := sources.Load(context.String("sources")); os.IsNotExist(err) {
			return fmt.Errorf("sources file is missing; run yae init to create it")
		} else {
			return err
		}
	}

	return &cli.App{
		Name:                 "yae",
		Version:              Version,
		Usage:                "Nix Dependency Manager",
		Description:          "Nix Dependency Manager",
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Fuwn",
				Email: "contact@fuwn.me",
			},
		},
		Before: func(context *cli.Context) error {
			log.SetLevel(log.InfoLevel)

			if context.Bool("debug") {
				log.SetLevel(log.DebugLevel)
			}

			if context.Bool("silent") {
				log.SetLevel(log.ErrorLevel)
			}

			return nil
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
			},
			&cli.BoolFlag{
				Name:  "silent",
				Usage: "Only log errors",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview changes without saving the sources file (downloads may still populate the Nix store)",
			},
		},
		Copyright: fmt.Sprintf("Copyright (c) 2024-%d Fuwn", time.Now().Year()),
		Suggest:   true,
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
				Before:    loadSources,
				Action:    commands.Add(&sources),
			},
			{
				Name:      "drop",
				ArgsUsage: "<name>",
				Args:      true,
				Usage:     "Drop a source",
				Before:    loadSources,
				Action:    commands.Drop(&sources),
			},
			{
				Name:      "update",
				Args:      true,
				Usage:     "Update one or all sources",
				ArgsUsage: "[name]",
				Flags:     commands.UpdateFlags(),
				Before:    loadSources,
				Action:    commands.Update(&sources),
			},
		},
	}
}
