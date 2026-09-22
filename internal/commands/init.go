package commands

import (
	"fmt"
	"os"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Init(sources *yae.Environment) func(context *cli.Context) error {
	return func(context *cli.Context) error {
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

		sources.Sources = make(map[string]yae.Source)
		sources.Schema = "https://raw.githubusercontent.com/Fuwn/yae/refs/heads/main/yae.schema.json"

		return sources.Save(context.String("sources"))
	}
}
