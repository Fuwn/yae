package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPackagedRuntime(t *testing.T) {
	binary := os.Getenv("YAE_TEST_BINARY")
	git := os.Getenv("YAE_TEST_GIT")

	if binary == "" || git == "" {
		t.Skip("set YAE_TEST_BINARY and YAE_TEST_GIT to test the installed package")
	}

	root := t.TempDir()
	sources := filepath.Join(root, "sources.json")
	payload := filepath.Join(root, "payload")
	repository := filepath.Join(root, "repository")

	for name, value := range map[string]string{
		"PATH": "", "HOME": root,
		"NIX_REMOTE":   "local?root=" + filepath.Join(root, "nix"),
		"NIX_CONF_DIR": filepath.Join(root, "config"), "NIX_USER_CONF_FILES": "",
		"NIX_CONFIG":          "experimental-features = nix-command\nbuild-users-group =\n",
		"GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.DevNull,
	} {
		t.Setenv(name, value)
	}

	run := func(executable string, arguments ...string) string {
		t.Helper()

		context, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		defer cancel()

		process := exec.CommandContext(context, executable, arguments...)

		var diagnostics bytes.Buffer

		process.Stderr = &diagnostics

		output, err := process.Output()

		if err != nil {
			t.Fatalf("%s %v: %v: %s", executable, arguments, err, diagnostics.String())
		}

		return strings.TrimSpace(string(output))
	}
	runYae := func(arguments ...string) string {
		return run(binary, append([]string{"--sources", sources}, arguments...)...)
	}
	write := func(path string, contents []byte) {
		t.Helper()

		if err := os.WriteFile(path, contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	runYae("init")
	write(payload, []byte("first payload"))
	runYae("add", "--type", "binary", "--unpack=false", "sample", (&url.URL{Scheme: "file", Path: payload}).String())

	contents, err := os.ReadFile(sources)

	if err != nil {
		t.Fatal(err)
	}

	var environment map[string]any

	if err := json.Unmarshal(contents, &environment); err != nil {
		t.Fatal(err)
	}

	source := environment["sample"].(map[string]any)
	digest := sha256.Sum256([]byte("first payload"))
	expected := "sha256-" + base64.StdEncoding.EncodeToString(digest[:])

	if source["hash"] != expected {
		t.Fatalf("saved hash = %v; want %s", source["hash"], expected)
	}

	run(git, "init", "--bare", repository)

	tree := run(git, "-C", repository, "mktree")
	commit := run(git, "-C", repository, "-c", "user.name=Yae Test", "-c", "user.email=yae@example.test", "commit-tree", tree, "-m", "fixture")

	run(git, "-C", repository, "update-ref", "refs/tags/v10", commit)
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "url."+(&url.URL{Scheme: "file", Path: repository}).String()+".insteadOf")
	t.Setenv("GIT_CONFIG_VALUE_0", "https://example.test/owner/repository")

	source["type"] = "git"
	source["version"] = "v10"
	source["url"] = "https://example.test/owner/repository/archive/v10"
	source["url_template"] = "https://example.test/owner/repository/archive/{version}"
	contents, err = json.Marshal(environment)

	if err != nil {
		t.Fatal(err)
	}

	write(sources, contents)

	if output := runYae("update", "--output-updated-list"); output != "" {
		t.Fatalf("packaged Git check changed a current source: %s", output)
	}
}
