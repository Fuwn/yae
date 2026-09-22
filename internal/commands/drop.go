package commands

import (
	"fmt"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Drop(sources *yae.Environment) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		if context.Args().Len() != 1 {
			return fmt.Errorf("drop requires exactly one source name")
		}

		if !sources.Exists(context.Args().Get(0)) {
			return fmt.Errorf("source does not exist")
		}

		if context.Bool("dry-run") {
			return nil
		}

		sources.Drop(context.Args().Get(0))

		return sources.Save(context.String("sources"))
	}
}
