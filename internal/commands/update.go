package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

func Update(sources *yae.Environment) *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update one or all sources",
		ArgsUsage: "[name]",
		Args:      true,
		Flags: []cli.Flag{
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
				Usage: "Refetch non-pinned git sources even when their version is unchanged",
			},
		},
		Before: loadSources(sources),
		Action: func(context *cli.Context) error {
			if context.Args().Len() > 1 {
				return fmt.Errorf("update accepts at most one source name")
			}

			names := context.Args().Slice()

			if len(names) == 0 {
				for name := range sources.Sources {
					names = append(names, name)
				}
			} else if _, exists := sources.Sources[names[0]]; !exists {
				return fmt.Errorf("source %q does not exist", names[0])
			}

			sort.Strings(names)

			updates := []string{}

			for _, name := range names {
				log.Infof("checking %s", name)

				source := sources.Sources[name]
				updated, err := source.Update(context.Context, context.Bool("force-hashed"))

				if err != nil {
					return fmt.Errorf("source %q: %w", name, err)
				}

				if updated != source {
					sources.Sources[name] = updated
					updates = append(updates, name)
				}
			}

			if len(updates) > 0 {
				if context.Bool("dry-run") {
					log.Infof("would update %s", strings.Join(updates, ", "))
				} else {
					if err := sources.Save(context.String("sources")); err != nil {
						return err
					}

					log.Infof("updated %s", strings.Join(updates, ", "))
				}
			}

			if context.Bool("output-updated-list") {
				for _, name := range updates {
					fmt.Fprintln(context.App.Writer, name)
				}
			} else if context.Bool("output-formatted-updated-list") {
				fmt.Fprintln(context.App.Writer, formatNames(updates))
			}

			return nil
		},
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
