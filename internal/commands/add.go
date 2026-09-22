package commands

import (
	"fmt"
	"strings"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func AddFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "unpack",
			Usage: "Unpack the source into the Nix Store",
			Value: true,
		},
		&cli.StringFlag{
			Name:     "type",
			Usage:    "Source type",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "version",
			Usage: "Source version used in identifying latest git source",
		},
		&cli.StringFlag{
			Name:  "tag-predicate",
			Usage: "Git tag predicate used in identifying latest git source",
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
	}
}

func Add(sources *yae.Environment) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Args().Len() != 2 {
			return fmt.Errorf("invalid number of arguments")
		}

		if sources.Exists(c.Args().Get(0)) {
			return fmt.Errorf("source already exists")
		}

		name := c.Args().Get(0)

		if strings.TrimSpace(name) == "" || name == "$schema" {
			return fmt.Errorf("source name must be non-empty and cannot be $schema")
		}

		source := yae.Source{
			URL:           c.Args().Get(1),
			Unpack:        c.Bool("unpack"),
			Type:          c.String("type"),
			Version:       c.String("version"),
			TagPredicate:  c.String("tag-predicate"),
			TrimTagPrefix: c.String("trim-tag-prefix"),
			Pinned:        c.Bool("pin"),
			Force:         c.Bool("force"),
		}

		if source.Version != "" {
			source.URLTemplate = source.URL
			source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", source.Version)
		}

		if err := source.Validate(); err != nil {
			return err
		}

		if err := source.RefreshHashes(c.Context); err != nil {
			return err
		}

		if err := sources.Add(name, source); err != nil {
			return err
		}

		if c.Bool("dry-run") {
			return nil
		}

		return sources.Save(c.String("sources"))
	}
}
