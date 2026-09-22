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

func TestHashOutputValidation(t *testing.T) {
	for _, output := range []string{"", "invalid", testSHA256 + "\nextra", testSHA256, testSHA256 + "\n"} {
		t.Run(output, func(t *testing.T) {
			fakeCommand(t, "nix-prefetch-url", "printf '%s' \"$YAE_TEST_HASH\"")
			t.Setenv("YAE_TEST_HASH", output)

			hash, err := fetchSHA256(context.Background(), "https://example.test/file", false)
			valid := strings.TrimSpace(output) == testSHA256

			if (err == nil) != valid || (valid && hash != testSHA256) {
				t.Fatalf("hash = %q, error = %v", hash, err)
			}
		})
	}
}

func TestFailedHashConversionLeavesSourceUnchanged(t *testing.T) {
	fakeNix(t)
	fakeCommand(t, "nix", "printf 'invalid'")

	source := validSource()
	original := source

	if err := source.RefreshHashes(context.Background()); err == nil || source != original {
		t.Fatalf("failed conversion changed source: %#v, %v", source, err)
	}
}

func TestFailedUpdateLeavesSourceUnchanged(t *testing.T) {
	fakeGit(t, "abc\trefs/tags/v2\n", false)
	fakeCommand(t, "nix-prefetch-url", "exit 1")

	source := Source{Type: "git", Version: "v1", URL: "https://example.test/owner/repo/archive/v1", URLTemplate: "https://example.test/owner/repo/archive/{version}", SHA256: testSHA256, Hash: testSRIHash}
	original := source

	if _, err := source.Update(context.Background(), false, false); err == nil || source != original {
		t.Fatalf("failed update changed source: %#v, %v", source, err)
	}
}

func TestForcedRehashDoesNotInventAChange(t *testing.T) {
	fakeGit(t, "abc\trefs/tags/v1\n", false)
	fakeNix(t)

	source := Source{Type: "git", Version: "v1", URL: "https://example.test/owner/repo/archive/v1", URLTemplate: "https://example.test/owner/repo/archive/{version}", SHA256: testSHA256, Hash: testSRIHash}
	updated, err := source.Update(context.Background(), true, false)

	if err != nil || updated != source {
		t.Fatalf("unchanged content reported as changed: %#v, %v", updated, err)
	}
}
