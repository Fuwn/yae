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
				ArgsUsage: "<name> <uri>",
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
						source.URITemplate = c.Args().Get(1)
						source.Version = c.String("version")

						if strings.Contains(source.URITemplate, "{version}") {
							source.URI = strings.Replace(source.URITemplate, "{version}", source.Version, 1)
						}
					} else {
						source.URI = c.Args().Get(1)
					}

					if source.Type == "git" && c.String("tag-predicate") != "" {
						source.TagPredicate = c.String("tag-predicate")
					}

					if sha256, err := fetchSHA256(source.URI, c.Bool("unpack")); err != nil {
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
				Action: func(c *cli.Context) error {
					if c.Args().Len() == 0 {
						for key, value := range sources {
							if err := updateSource(&sources, key, value); err != nil {
								return err
							}
						}
					} else {
						name := c.Args().Get(0)

						if err := updateSource(&sources, name, sources[name]); err != nil {
							return err
						}
					}

					if err := sources.Save(c.String("sources")); err != nil {
						return err
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
	executable, err := exec.LookPath(name)

	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()

	return string(out), err
}

func fetchLatestGitTag(source Source) (string, error) {
	if source.Type == "git" {
		repository := "https://github.com/" + strings.Split(source.URI, "/")[3] + "/" + strings.Split(source.URI, "/")[4]
		remotes, err := commandOutput("git", "ls-remote", "--tags", repository)

		if err != nil {
			return "", err
		}

		refs := strings.Split(remotes, "\n")
		var latest string

		if source.TagPredicate == "" {
			latest = refs[len(refs)-2]
		} else {
			for i := len(refs) - 2; i >= 0; i-- {
				if strings.Contains(refs[i], source.TagPredicate) {
					latest = strings.Split(refs[i], "/")[2]

					break
				}
			}
		}

		return latest, nil
	}

	return "", fmt.Errorf("source is not a git repository")
}

func updateSource(sources *Sources, name string, source Source) error {
	if !sources.Exists(name) {
		return fmt.Errorf("source does not exist")
	}

	if source.Type == "git" {
		tag, err := fetchLatestGitTag(source)

		if err != nil {
			return err
		}

		if tag != source.Version {
			fmt.Println("updated version for", name, "from", source.Version, "to", tag)

			source.Version = tag

			if strings.Contains(source.URITemplate, "{version}") {
				source.URI = strings.Replace(source.URITemplate, "{version}", source.Version, 1)
			}
		}
	}

	sha256, err := fetchSHA256(source.URI, source.Unpack)

	if err != nil {
		return err
	}

	if sha256 != source.SHA256 {
		fmt.Println("updated hash for", name, "from", source.SHA256, "to", sha256)

		source.SHA256 = sha256
	}

	(*sources)[name] = source

	return nil
}
