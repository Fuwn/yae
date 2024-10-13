package commands

import (
	"fmt"
	"strings"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Add(sources *yae.Sources) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Args().Len() != 2 {
			return fmt.Errorf("invalid number of arguments")
		}

		if sources.Exists(c.Args().Get(0)) {
			return fmt.Errorf("source already exists")
		}

		source := yae.Source{
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

		if sha256, err := yae.FetchSHA256(source.URL, c.Bool("unpack")); err != nil {
			return err
		} else {
			source.SHA256 = sha256
		}

		if err := sources.Add(c.Args().Get(0), source); err != nil {
			return err
		}

		return sources.Save(c.String("sources"))
	}
}
