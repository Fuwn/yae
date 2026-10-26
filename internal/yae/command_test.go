package yae

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fakeCommand(t *testing.T, name string, body string) {
	t.Helper()

	shell, err := exec.LookPath("sh")

	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()

	if err := os.WriteFile(filepath.Join(directory, name), []byte("#!"+shell+"\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestCommandFailureIncludesDiagnostics(t *testing.T) {
	fakeCommand(t, "git", "printf 'remote unavailable' >&2\nexit 23")

	_, err := command(context.Background(), "git", "ls-remote")

	if err == nil || !strings.Contains(err.Error(), "git") || !strings.Contains(err.Error(), "remote unavailable") || !strings.Contains(err.Error(), "23") {
		t.Fatalf("missing diagnostics: %v", err)
	}
}

func TestCommandCancellation(t *testing.T) {
	fakeCommand(t, "git", "while :; do :; done")

	context, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

	defer cancel()

	started := time.Now()
	_, err := command(context, "git")

	if !errors.Is(err, context.Err()) || time.Since(started) > 5*time.Second {
		t.Fatalf("command did not cancel promptly: %v", err)
	}
}
