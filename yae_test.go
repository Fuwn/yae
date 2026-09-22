package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Fuwn/yae/internal/yae"
	"github.com/charmbracelet/log"
)

const fixtureSHA256 = "1wn29537l343lb0id0byk0699fj0k07m1n2d7jx2n0ssax55vhwy"
const fixtureSRIHash = "sha256-nsNdSldaAyu6PE3YUA+YQLqUDJh+gRbBooMMekZJwvI="

func runCLI(t *testing.T, path string, arguments ...string) (string, string, error) {
	t.Helper()

	var output, diagnostics bytes.Buffer

	log.SetOutput(&diagnostics)

	defer log.SetOutput(os.Stderr)

	application := newApp()

	application.Writer = &output
	application.ErrWriter = &diagnostics

	err := application.RunContext(context.Background(), append([]string{"yae", "--sources", path}, arguments...))

	return output.String(), diagnostics.String(), err
}

func installFixtureCommands(t *testing.T) {
	t.Helper()

	shell, err := exec.LookPath("sh")

	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	commands := map[string]string{
		"git":              "printf 'abc\\trefs/tags/v2\\n'",
		"nix":              "printf '%s\\n' '" + fixtureSRIHash + "'",
		"nix-prefetch-url": "case \"$*\" in *fail*) printf 'fetch failed' >&2; exit 23;; esac\nprintf '%s\\n' '" + fixtureSHA256 + "'",
	}

	for name, body := range commands {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("#!"+shell+"\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func readSources(t *testing.T, path string) yae.Environment {
	t.Helper()

	var environment yae.Environment

	if err := environment.Load(path); err != nil {
		t.Fatal(err)
	}

	return environment
}

func TestHelpWithoutEnvironment(t *testing.T) {
	for _, arguments := range [][]string{{}, {"--help"}, {"help", "add"}, {"init", "--help"}, {"add", "--help"}, {"update", "--help"}, {"drop", "--help"}} {
		output, _, err := runCLI(t, filepath.Join(t.TempDir(), "missing.json"), arguments...)

		if err != nil || !strings.Contains(output, "USAGE:") {
			t.Fatalf("help %v failed: %s, %v", arguments, output, err)
		}
	}
}

func TestCommandLifecycle(t *testing.T) {
	installFixtureCommands(t)

	path := filepath.Join(t.TempDir(), "sources.json")

	for _, arguments := range [][]string{
		{"init"},
		{"add", "--type", "binary", "binary", "https://example.test/file"},
		{"add", "--type", "git", "--version", "v1", "release", "https://example.test/owner/repo/archive/{version}"},
	} {
		if _, _, err := runCLI(t, path, arguments...); err != nil {
			t.Fatal(err)
		}
	}

	environment := readSources(t, path)

	for name, source := range environment.Sources {
		if source.SHA256 != fixtureSHA256 || source.Hash != fixtureSRIHash {
			t.Fatalf("%s was added with incomplete hashes: %#v", name, source)
		}
	}

	output, _, err := runCLI(t, path, "update", "--output-updated-list", "release")

	if err != nil || output != "release\n" {
		t.Fatalf("update output = %q, error = %v", output, err)
	}

	environment = readSources(t, path)

	if environment.Sources["release"].Version != "v2" || !strings.HasSuffix(environment.Sources["release"].URL, "/v2") {
		t.Fatalf("release did not advance: %#v", environment.Sources["release"])
	}

	if _, _, err := runCLI(t, path, "drop", "binary"); err != nil {
		t.Fatal(err)
	}

	if readSources(t, path).Exists("binary") {
		t.Fatal("dropped source remains")
	}
}

func TestInvalidArgumentsPreserveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.json")

	if _, _, err := runCLI(t, path, "init"); err != nil {
		t.Fatal(err)
	}

	original, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	for _, arguments := range [][]string{
		{"init"}, {"init", "extra"}, {"drop"}, {"drop", "one", "two"},
		{"update", "one", "two"}, {"update", "absent"}, {"drop", "absent"},
		{"add", "--type", "binary", "$schema", "https://example.test/file"},
		{"add", "--type", "git", "missing-version", "https://example.test/owner/repo/archive/v1"},
	} {
		if _, _, err := runCLI(t, path, arguments...); err == nil {
			t.Fatalf("accepted invalid arguments: %v", arguments)
		}

		data, err := os.ReadFile(path)

		if err != nil || !bytes.Equal(data, original) {
			t.Fatalf("invalid command %v changed file: %v", arguments, err)
		}
	}
}

func TestDryRunPreservesFiles(t *testing.T) {
	installFixtureCommands(t)

	path := filepath.Join(t.TempDir(), "sources.json")

	if _, _, err := runCLI(t, path, "--dry-run", "init"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("dry-run init created a file")
	}

	if _, _, err := runCLI(t, path, "init"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runCLI(t, path, "add", "--type", "git", "--version", "v1", "release", "https://example.test/owner/repo/archive/{version}"); err != nil {
		t.Fatal(err)
	}

	original, err := os.ReadFile(path)

	if err != nil {
		t.Fatal(err)
	}

	for _, arguments := range [][]string{{"drop", "release"}, {"update"}, {"add", "--type", "binary", "new", "https://example.test/file"}} {
		if _, _, err := runCLI(t, path, append([]string{"--dry-run"}, arguments...)...); err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(path)

		if err != nil || !bytes.Equal(data, original) {
			t.Fatalf("dry-run %v changed file: %v", arguments, err)
		}
	}
}

func TestBooleanLoggingFlags(t *testing.T) {
	installFixtureCommands(t)

	path := filepath.Join(t.TempDir(), "sources.json")

	if _, _, err := runCLI(t, path, "init"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runCLI(t, path, "add", "--type", "binary", "sample", "https://example.test/file"); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		flag  string
		info  bool
		debug bool
	}{
		{flag: "--debug=false", info: true},
		{flag: "--silent=false", info: true},
		{flag: "--debug", info: true, debug: true},
		{flag: "--silent"},
	}

	for _, test := range cases {
		_, diagnostics, err := runCLI(t, path, test.flag, "update")

		if err != nil || strings.Contains(diagnostics, "checking sample") != test.info || strings.Contains(diagnostics, "running nix") != test.debug {
			t.Fatalf("%s: diagnostics = %q, error = %v", test.flag, diagnostics, err)
		}
	}
}

func TestUpdateIsOrderedAndFailureDoesNotSave(t *testing.T) {
	installFixtureCommands(t)

	for _, failing := range []bool{false, true} {
		path := filepath.Join(t.TempDir(), "sources.json")
		source := yae.Source{URL: "https://example.test/file", Type: "binary", SHA256: strings.Repeat("0", 52), Hash: "sha256-" + strings.Repeat("A", 43) + "="}
		environment := yae.Environment{Sources: map[string]yae.Source{"zulu": source, "alpha": source}}

		if failing {
			source.URL = "https://example.test/fail"
			environment.Sources["zulu"] = source
		}

		if err := environment.Save(path); err != nil {
			t.Fatal(err)
		}

		original, err := os.ReadFile(path)

		if err != nil {
			t.Fatal(err)
		}

		output, diagnostics, err := runCLI(t, path, "update", "--output-updated-list")

		if failing {
			data, readError := os.ReadFile(path)

			if err == nil || output != "" || strings.Contains(diagnostics, "updated ") || readError != nil || !bytes.Equal(original, data) {
				t.Fatalf("failed update published changes: %q, %q, %v, %v", output, diagnostics, err, readError)
			}
		} else if err != nil || output != "alpha\nzulu\n" {
			t.Fatalf("unordered update: %q, %v", output, err)
		}
	}
}

func TestPinnedLegacyHashRepair(t *testing.T) {
	installFixtureCommands(t)

	path := filepath.Join(t.TempDir(), "sources.json")
	source := yae.Source{URL: "https://example.test/file", Type: "binary", SHA256: fixtureSHA256, Pinned: true}
	data, err := json.Marshal(map[string]yae.Source{"pinned": source})

	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := runCLI(t, path, "update"); err != nil {
		t.Fatal(err)
	}

	updated := readSources(t, path).Sources["pinned"]

	if updated.Hash != fixtureSRIHash || updated.SHA256 != source.SHA256 || !updated.Pinned || updated.URL != source.URL {
		t.Fatalf("repair changed pinned source: %#v", updated)
	}
}
