package commands

import (
	"fmt"
	"os"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func Init(sources *yae.Sources) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if _, err := os.Stat(c.String("sources")); err == nil {
			return fmt.Errorf("sources file already exists")
		}

		return sources.Save(c.String("sources"))
	}
}
