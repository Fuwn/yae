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
				},
				Action: func(c *cli.Context) error {
					showAll := !c.Bool("show-updated-only") && !c.Bool("show-updated-only-formatted")
					updates := []string{}

					if c.Args().Len() == 0 {
						for name, value := range sources {
							if updated, err := updateSource(&sources, name, value, showAll); err != nil {
								return err
							} else if updated {
								updates = append(updates, name)
							}
						}
					} else {
						name := c.Args().Get(0)

						if updated, err := updateSource(&sources, name, sources[name], showAll); err != nil {
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

func fetchSHA256(url string, unpack bool, show bool) (string, error) {
	arguments := []string{"--type", "sha256", url}

	if unpack {
		arguments = append([]string{"--unpack"}, arguments...)
	}

	output, err := command("nix-prefetch-url", show, arguments...)

	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")

	return strings.Trim(lines[len(lines)-2], "\n"), nil
}

func command(name string, show bool, args ...string) (string, error) {
	executable, err := exec.LookPath(name)
	out := []byte{}

	if show {
		cmd := exec.Command(executable, args...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err = cmd.Output()
	} else {
		cmd := exec.Command(executable, args...)
		out, err = cmd.Output()
	}

	return string(out), err
}

func fetchLatestGitTag(source Source, show bool) (string, error) {
	if source.Type == "git" {
		repository := "https://github.com/" + strings.Split(source.URL, "/")[3] + "/" + strings.Split(source.URL, "/")[4]
		remotes, err := command("bash", show, "-c", fmt.Sprintf("git ls-remote --tags %s | awk -F'/' '{print $NF}' | sort -V", repository))

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
					latest = refs[i]

					break
				}
			}
		}

		if source.TrimTagPrefix != "" {
			latest = strings.TrimPrefix(latest, source.TrimTagPrefix)
		}

		return latest, nil
	}

	return "", fmt.Errorf("source is not a git repository")
}

func updateSource(sources *Sources, name string, source Source, show bool) (bool, error) {
	updated := false

	if !sources.Exists(name) {
		return updated, fmt.Errorf("source does not exist")
	}

	if source.Pinned {
		if show {
			fmt.Println("skipped update for", name, "because it is pinned")
		}

		return updated, nil
	}

	if source.Type == "git" {
		tag, err := fetchLatestGitTag(source, show)

		if err != nil {
			return updated, err
		}

		if tag != source.Version {
			if show {
				fmt.Println("updated version for", name, "from", source.Version, "to", tag)
			}

			source.Version = tag
			updated = true

			if strings.Contains(source.URLTemplate, "{version}") {
				source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", source.Version)
			}
		}
	}

	sha256, err := fetchSHA256(source.URL, source.Unpack, show)

	if err != nil {
		return updated, err
	}

	if sha256 != source.SHA256 {
		if show {
			fmt.Println("updated hash for", name, "from", source.SHA256, "to", sha256)
		}

		source.SHA256 = sha256
		updated = true
	}

	(*sources)[name] = source

	return updated, nil
}

func lister(items []string) string {
	if len(items) == 0 {
		return ""
	} else if len(items) == 1 {
		return items[0]
	} else if len(items) == 2 {
		return fmt.Sprintf("%s & %s", items[0], items[1])
	}

	return fmt.Sprintf("%s, & %s", strings.Join(items[:len(items)-1], ", "), items[len(items)-1])
}
