package commands

import (
	"fmt"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Drop(sources *yae.Sources) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Args().Len() == 0 {
			return fmt.Errorf("invalid number of arguments")
		}

		if !sources.Exists(c.Args().Get(0)) {
			return fmt.Errorf("source does not exist")
		}

		sources.Drop(c.Args().Get(0))

		if c.Bool("dry-run") {
			return nil
		}

		return sources.Save(c.String("sources"))
	}
}
