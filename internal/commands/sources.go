package commands

import (
	"fmt"
	"os"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/urfave/cli/v2"
)

func loadSources(sources *yae.Environment) cli.BeforeFunc {
	return func(context *cli.Context) error {
		err := sources.Load(context.String("sources"))

		if os.IsNotExist(err) {
			return fmt.Errorf("sources file is missing; run yae init to create it")
		}

		return err
	}
}
