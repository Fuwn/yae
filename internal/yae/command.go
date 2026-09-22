package yae

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

func command(context context.Context, name string, arguments ...string) (string, error) {
	process := exec.CommandContext(context, name, arguments...)

	var stderr bytes.Buffer

	process.Stderr = &stderr
	process.WaitDelay = time.Second

	if name == "git" {
		process.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	}

	log.Debugf("running %s", name)

	output, err := process.Output()

	if context.Err() != nil {
		return "", fmt.Errorf("%s: %w", name, context.Err())
	}

	if err != nil {
		return "", fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}

	return string(output), nil
}
