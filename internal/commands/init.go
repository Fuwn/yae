package commands

import (
	"fmt"
	"os"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Init(sources *yae.Environment) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if _, err := os.Stat(c.String("sources")); err == nil {
			return fmt.Errorf("sources file already exists")
		}

		sources.Sources = make(map[string]yae.Source)
		sources.Schema = "https://raw.githubusercontent.com/Fuwn/yae/refs/heads/main/yae.schema.json"

		if c.Bool("dry-run") {
			return nil
		}

		return sources.Save(c.String("sources"))
	}
}
