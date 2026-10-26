package yae

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Fuwn/yae/internal/fixture"
)

func TestCommandFailureIncludesDiagnostics(t *testing.T) {
	fixture.Command(t, "git", "printf 'remote unavailable' >&2\nexit 23")

	_, err := command(context.Background(), "git", "ls-remote")

	if err == nil || !strings.Contains(err.Error(), "git") || !strings.Contains(err.Error(), "remote unavailable") || !strings.Contains(err.Error(), "23") {
		t.Fatalf("missing diagnostics: %v", err)
	}
}

func TestCommandCancellation(t *testing.T) {
	fixture.Command(t, "git", "while :; do :; done")

	context, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

	defer cancel()

	started := time.Now()
	_, err := command(context, "git")

	if !errors.Is(err, context.Err()) || time.Since(started) > 5*time.Second {
		t.Fatalf("command did not cancel promptly: %v", err)
	}
}
