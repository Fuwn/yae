package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	sources := Sources{}

	(&cli.App{
		Name:                 "wiene",
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
			return sources.Load(c.String("sources"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "sources",
				Value: "./sources.json",
				Usage: "Sources path",
			},
		},
		Copyright: fmt.Sprintf("Copyright (c) 2024-%s Fuwn", fmt.Sprint(time.Now().Year())),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				fmt.Println(err)
			}
		},
		Suggest: true,
		Commands: []*cli.Command{
			{
				Name:      "update",
				Args:      true,
				Usage:     "Update one or all sources",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "unpack",
						Usage: "Unpack the source into the Nix Store",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Args().Len() == 0 {
						for key, value := range sources {
							sha256, err := fetchSHA256(value.Url, value.Unpack)

							if err != nil {
								return err
							}

							if sha256 != value.SHA256 {
								sources[key] = Source{
									Url:    value.Url,
									SHA256: sha256,
									Unpack: value.Unpack,
								}

								fmt.Println("updated hash for", key)
							}

							if err = sources.Save(c.String("sources")); err != nil {
								return err
							}
						}
					} else {
						if !sources.Exists(c.Args().Get(0)) {
							return fmt.Errorf("source does not exist")
						}

						sha256, err := fetchSHA256(sources[c.Args().Get(0)].Url, c.Bool("unpack"))

						if err != nil {
							return err
						}

						sources[c.Args().Get(0)] = Source{
							Url:    sources[c.Args().Get(0)].Url,
							SHA256: sha256,
							Unpack: c.Bool("unpack"),
						}

						if err = sources.Save(c.String("sources")); err != nil {
							return err
						}
					}

					return nil
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
				Name:      "add",
				Args:      true,
				ArgsUsage: "<name> <uri>",
				Usage:     "Add a source",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "unpack",
						Usage: "Unpack the source into the Nix Store",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Args().Len() != 2 {
						return fmt.Errorf("invalid number of arguments")
					}

					if sources.Exists(c.Args().Get(0)) {
						return fmt.Errorf("source already exists")
					}

					sha256, err := fetchSHA256(c.Args().Get(1), c.Bool("unpack"))

					if err != nil {
						return err
					}

					if err = sources.Add(c.Args().Get(0), Source{
						Url:    c.Args().Get(1),
						SHA256: sha256,
						Unpack: c.Bool("unpack"),
					}); err != nil {
						return err
					}

					return sources.Save(c.String("sources"))
				},
			},
		},
	}).Run(os.Args)
}

func fetchSHA256(uri string, unpack bool) (string, error) {
	arguments := []string{"--type", "sha256", uri}

	if unpack {
		arguments = append([]string{"--unpack"}, arguments...)
	}

	output, err := commandOutput("nix-prefetch-url", arguments...)

	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")

	return strings.Trim(lines[len(lines)-2], "\n"), nil
}

func commandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()

	return string(out), err
}
