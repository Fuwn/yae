package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	sources := Sources{}

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
		},
		Copyright: fmt.Sprintf("Copyright (c) 2024-%s Fuwn", fmt.Sprint(time.Now().Year())),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
		Suggest: true,
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialise a new Yae environment",
				Action: func(c *cli.Context) error {
					if _, err := os.Stat(c.String("sources")); err == nil {
						return fmt.Errorf("sources file already exists")
					}

					return sources.Save(c.String("sources"))
				},
			},
			{
				Name:      "add",
				Args:      true,
				ArgsUsage: "<name> <url>",
				Usage:     "Add a source",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "unpack",
						Usage: "Unpack the source into the Nix Store",
						Value: true,
					},
					&cli.StringFlag{
						Name:     "type",
						Usage:    "Source type",
						Required: true,
						Action: func(c *cli.Context, value string) error {
							if value != "binary" && value != "git" {
								return fmt.Errorf("invalid source type: must be 'binary' or 'git'")
							}

							return nil
						},
					},
					&cli.StringFlag{
						Name:  "version",
						Usage: "Source version used in identifying latest git source",
					},
					&cli.StringFlag{
						Name:  "tag-predicate",
						Usage: "Git tag predicate used in identifying latest git source",
					},
					&cli.BoolFlag{
						Name:  "silent",
						Usage: "Silence output",
					},
					&cli.StringFlag{
						Name:  "trim-tag-prefix",
						Usage: "A prefix to trim from remote git tags",
					},
					&cli.BoolFlag{
						Name:  "pin",
						Usage: "Prevent the source from being updated",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Always force update the source, regardless of unchanged remote tag",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Args().Len() != 2 {
						return fmt.Errorf("invalid number of arguments")
					}

					if sources.Exists(c.Args().Get(0)) {
						return fmt.Errorf("source already exists")
					}

					source := Source{
						Unpack: c.Bool("unpack"),
						Type:   c.String("type"),
					}
					version := c.String("version")

					if version != "" {
						source.URLTemplate = c.Args().Get(1)
						source.Version = c.String("version")

						if strings.Contains(source.URLTemplate, "{version}") {
							source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", source.Version)
						}
					} else {
						source.URL = c.Args().Get(1)
					}

					if source.Type == "git" && c.String("tag-predicate") != "" {
						source.TagPredicate = c.String("tag-predicate")
					}

					if c.String("trim-tag-prefix") != "" {
						source.TrimTagPrefix = c.String("trim-tag-prefix")
					}

					if c.Bool("pin") {
						source.Pinned = true
					}

					if c.Bool("force") {
						if source.Pinned {
							return fmt.Errorf("cannot set a source to be statically forced and pinned at the same time")
						}

						source.Force = true
					}

					if sha256, err := fetchSHA256(source.URL, c.Bool("unpack"), !c.Bool("silent")); err != nil {
						return err
					} else {
						source.SHA256 = sha256
					}

					if err := sources.Add(c.Args().Get(0), source); err != nil {
						return err
					}

					return sources.Save(c.String("sources"))
				},
			},
			{
				Name:  "drop",
				Args:  true,
				Usage: "Drop a source",
				Action: func(c *cli.Context) error {
					if c.Args().Len() == 0 {
						return fmt.Errorf("invalid number of arguments")
					}

					if !sources.Exists(c.Args().Get(0)) {
						return fmt.Errorf("source does not exist")
					}

					sources.Drop(c.Args().Get(0))

					return sources.Save(c.String("sources"))
				},
			},
			{
				Name:      "update",
				Args:      true,
				Usage:     "Update one or all sources",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "show-updated-only",
						Usage: "Output a newline-seperated list of updated sources, silence other output",
					},
					&cli.BoolFlag{
						Name:  "show-updated-only-formatted",
						Usage: "Output a comma and/or ampersand list of updated sources, silence other output",
					},
					&cli.BoolFlag{
						Name:  "force-hashed",
						Usage: "Force updates for non-pinned sources that have an unchanged version (recalculate hash)",
					},
					&cli.BoolFlag{
						Name:  "force-pinned",
						Usage: "Force updates for all sources, including pinned sources (can be used with --force-hashed)",
					},
				},
				Action: func(c *cli.Context) error {
					showAll := !c.Bool("show-updated-only") && !c.Bool("show-updated-only-formatted")
					updates := []string{}
					force := c.Bool("force-hashed")
					forcePinned := c.Bool("force-pinned")

					if c.Args().Len() == 0 {
						for name, source := range sources {
							if updated, err := source.Update(&sources, name, showAll, force, forcePinned); err != nil {
								return err
							} else if updated {
								updates = append(updates, name)
							}
						}
					} else {
						name := c.Args().Get(0)
						source := sources[name]

						if updated, err := source.Update(&sources, name, showAll, force, forcePinned); err != nil {
							return err
						} else if updated {
							updates = append(updates, name)
						}
					}

					if len(updates) > 0 {
						if err := sources.Save(c.String("sources")); err != nil {
							return err
						}
					}

					if c.Bool("show-updated-only") {
						for _, update := range updates {
							fmt.Println(update)
						}
					} else if c.Bool("show-updated-only-formatted") {
						fmt.Println(lister(updates))
					}

					return nil
				},
			},
		},
	}).Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
