package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

func UpdateFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "output-updated-list",
			Usage: "Output a newline-separated list of updated sources, regardless of silent mode",
		},
		&cli.BoolFlag{
			Name:  "output-formatted-updated-list",
			Usage: "Output a comma and/or ampersand list of updated sources, regardless of silent mode",
		},
		&cli.BoolFlag{
			Name:  "force-hashed",
			Usage: "Force updates for non-pinned sources that have an unchanged version (recalculate hash)",
		},
		&cli.BoolFlag{
			Name:  "force-pinned",
			Usage: "Force updates for all sources, including pinned sources (can be used with --force-hashed)",
		},
	}
}

func Update(sources *yae.Environment) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Args().Len() > 1 {
			return fmt.Errorf("update accepts at most one source name")
		}

		names := c.Args().Slice()

		if len(names) == 0 {
			for name := range sources.Sources {
				names = append(names, name)
			}
		} else if !sources.Exists(names[0]) {
			return fmt.Errorf("source %q does not exist", names[0])
		}

		sort.Strings(names)

		updates := []string{}
		pending := yae.Environment{Schema: sources.Schema, Sources: make(map[string]yae.Source, len(sources.Sources))}

		for name, source := range sources.Sources {
			pending.Sources[name] = source
		}

		for _, name := range names {
			log.Infof("checking %s", name)

			source := sources.Sources[name]
			updated, err := source.Update(c.Context, c.Bool("force-hashed"), c.Bool("force-pinned"))

			if err != nil {
				return fmt.Errorf("source %q: %w", name, err)
			}

			if updated != source {
				pending.Sources[name] = updated
				updates = append(updates, name)
			}
		}

		if len(updates) > 0 {
			if c.Bool("dry-run") {
				log.Infof("would update %s", strings.Join(updates, ", "))
			} else {
				if err := pending.Save(c.String("sources")); err != nil {
					return err
				}

				*sources = pending

				log.Infof("updated %s", strings.Join(updates, ", "))
			}
		}

		if c.Bool("output-updated-list") {
			for _, name := range updates {
				fmt.Fprintln(c.App.Writer, name)
			}
		} else if c.Bool("output-formatted-updated-list") {
			fmt.Fprintln(c.App.Writer, formatNames(updates))
		}

		return nil
	}
}

func formatNames(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}

	if len(names) == 2 {
		return strings.Join(names, " & ")
	}

	return strings.Join(names[:len(names)-1], ", ") + ", & " + names[len(names)-1]
}
