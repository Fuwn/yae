package commands

import (
	"fmt"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Drop(sources *yae.Environment) *cli.Command {
	return &cli.Command{
		Name:      "drop",
		Usage:     "Drop a source",
		ArgsUsage: "<name>",
		Args:      true,
		Before:    loadSources(sources),
		Action: func(context *cli.Context) error {
			if context.Args().Len() != 1 {
				return fmt.Errorf("drop requires exactly one source name")
			}

			name := context.Args().First()

			if _, exists := sources.Sources[name]; !exists {
				return fmt.Errorf("source %q does not exist", name)
			}

			if context.Bool("dry-run") {
				return nil
			}

			delete(sources.Sources, name)

			return sources.Save(context.String("sources"))
		},
	}
}
