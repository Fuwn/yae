package commands

import (
	"fmt"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func UpdateFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "output-updated-list",
			Usage: "Output a newline-seperated list of updated sources, regardless of silent mode",
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

func Update(sources *yae.Sources) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		updates := []string{}
		force := c.Bool("force-hashed")
		forcePinned := c.Bool("force-pinned")

		if c.Args().Len() == 0 {
			for name, source := range *sources {
				if updated, err := source.Update(sources, name, force, forcePinned); err != nil {
					return err
				} else if updated {
					updates = append(updates, name)
				}
			}
		} else {
			name := c.Args().Get(0)
			source := (*sources)[name]

			if updated, err := source.Update(sources, name, force, forcePinned); err != nil {
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

		if c.Bool("output-updated-list") {
			for _, update := range updates {
				fmt.Println(update)
			}
		} else if c.Bool("output-formatted-updated-list") {
			fmt.Println(yae.Lister(updates))
		}

		return nil
	}
}
