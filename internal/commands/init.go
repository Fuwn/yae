package commands

import (
	"fmt"
	"os"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Init() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialise a new Yae environment",
		Action: func(context *cli.Context) error {
			if context.Args().Len() != 0 {
				return fmt.Errorf("init does not accept arguments")
			}

			if _, err := os.Lstat(context.String("sources")); err == nil {
				return fmt.Errorf("sources file already exists")
			} else if !os.IsNotExist(err) {
				return err
			}

			if context.Bool("dry-run") {
				return nil
			}

			sources := yae.Environment{
				Schema:  "https://raw.githubusercontent.com/Fuwn/yae/refs/heads/main/yae.schema.json",
				Sources: map[string]yae.Source{},
			}

			return sources.Save(context.String("sources"))
		},
	}
}
