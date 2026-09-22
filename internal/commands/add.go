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

func Add(sources *yae.Environment) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		if context.Args().Len() != 2 {
			return fmt.Errorf("invalid number of arguments")
		}

		name := context.Args().Get(0)

		if err := sources.CheckNewName(name); err != nil {
			return err
		}

		source := yae.Source{
			URL:           context.Args().Get(1),
			Unpack:        context.Bool("unpack"),
			Type:          context.String("type"),
			Version:       context.String("version"),
			TagPredicate:  context.String("tag-predicate"),
			TrimTagPrefix: context.String("trim-tag-prefix"),
			Pinned:        context.Bool("pin"),
			Force:         context.Bool("force"),
		}

		if source.Version != "" {
			source.URLTemplate = source.URL
			source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", source.Version)
		}

		if err := source.Validate(); err != nil {
			return err
		}

		if err := source.RefreshHashes(context.Context); err != nil {
			return err
		}

		if context.Bool("dry-run") {
			return nil
		}

		if err := sources.Add(name, source); err != nil {
			return err
		}

		return sources.Save(context.String("sources"))
	}
}
