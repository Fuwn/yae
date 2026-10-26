package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fuwn/yae/internal/commands"
	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

var Version string

func main() {
	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := newApp().RunContext(context, os.Args)

	stop()

	if err != nil {
		fmt.Fprintln(os.Stderr, "yae:", err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	sources := yae.Environment{}

	return &cli.App{
		Name:                 "yae",
		Version:              Version,
		Usage:                "Nix Dependency Manager",
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Fuwn",
				Email: "contact@fuwn.me",
			},
		},
		Before: func(context *cli.Context) error {
			level := slog.LevelInfo

			if context.Bool("debug") {
				level = slog.LevelDebug
			}

			if context.Bool("silent") {
				level = slog.LevelError
			}

			handler := slog.NewTextHandler(context.App.ErrWriter, &slog.HandlerOptions{Level: level, ReplaceAttr: omitTime})

			slog.SetDefault(slog.New(handler))

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
			commands.Init(),
			commands.Add(&sources),
			commands.Drop(&sources),
			commands.Update(&sources),
		},
	}
}

func omitTime(groups []string, attribute slog.Attr) slog.Attr {
	if len(groups) == 0 && attribute.Key == slog.TimeKey {
		return slog.Attr{}
	}

	return attribute
}
