package yae

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func fakeGit(t *testing.T, output string, failure bool) string {
	t.Helper()

	shell, err := exec.LookPath("sh")

	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	arguments := filepath.Join(directory, "arguments")
	script := "#!" + shell + "\nprintf '%s\\n' \"$@\" > \"$YAE_TEST_ARGUMENTS\"\nprintf '%s' \"$YAE_TEST_TAGS\"\n"

	if failure {
		script += "exit 23\n"
	}

	if err := os.WriteFile(filepath.Join(directory, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("YAE_TEST_ARGUMENTS", arguments)
	t.Setenv("YAE_TEST_TAGS", output)

	return arguments
}

func TestLatestGitTag(t *testing.T) {
	cases := []struct {
		name      string
		output    string
		predicate string
		prefix    string
		failure   bool
		want      string
	}{
		{name: "tags only", output: "abc\trefs/heads/zzbranch\nabc\trefs/tags/v10\nabc\trefs/tags/v2\n", want: "v10"},
		{name: "annotated tags", output: "abc\trefs/tags/v10^{}\nabc\trefs/tags/v10", want: "v10"},
		{name: "nested tag", output: "abc\trefs/tags/release/v2\n", want: "release/v2"},
		{name: "filter before trimming", output: "abc\trefs/tags/v2-beta\nabc\trefs/tags/v1\n", predicate: "^v[0-9]+$", prefix: "v", want: "1"},
		{name: "no tags"},
		{name: "no match", output: "abc\trefs/tags/v1\n", predicate: "^never$"},
		{name: "invalid regex", predicate: "["},
		{name: "empty trimmed tag", output: "abc\trefs/tags/v1\n", prefix: "v1"},
		{name: "remote failure", failure: true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fakeGit(t, test.output, test.failure)

			source := Source{Type: "git", URL: "https://example.test/owner/repo/archive/v0.tar.gz", TagPredicate: test.predicate, TrimTagPrefix: test.prefix}
			version, err := source.fetchLatestGitTag()

			if test.want == "" {
				if err == nil {
					t.Fatalf("expected an error, got %q", version)
				}

				return
			}

			if err != nil || version != test.want {
				t.Fatalf("version = %q, error = %v; want %q", version, err, test.want)
			}
		})
	}
}

func TestGitURLIsAnArgument(t *testing.T) {
	arguments := fakeGit(t, "abc\trefs/tags/v1\n", false)
	source := Source{Type: "git", URL: "https://example.test/$(printf${IFS}INJECTED)/repo"}

	if _, err := source.fetchLatestGitTag(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(arguments)

	if err != nil {
		t.Fatal(err)
	}

	repository, err := repositoryURL(source.URL)

	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "ls-remote\n--tags\n--refs\n--sort=-version:refname\n--\n"+repository+"\n" || strings.Contains(string(data), "/INJECTED/repo") {
		t.Fatalf("unexpected git arguments: %s", data)
	}
}

func TestRepositoryURL(t *testing.T) {
	cases := []struct{ address, want string }{
		{"https://github.com/owner/repo/releases/download/v1/file", "https://github.com/owner/repo"},
		{"https://gitlab.com/group/nested/repo/-/archive/v1/file", "https://gitlab.com/group/nested/repo"},
		{"https://example.test/owner/repo/archive/v1.tar.gz?download=1", "https://example.test/owner/repo"},
		{"https://example.test", ""},
		{"https://example.test/owner", ""},
		{"https://example.test/owner/repo/unknown/file", ""},
		{"https://example.test/owner//archive/file", ""},
		{"https://example.test/../repo/archive/file", ""},
		{"file:///owner/repo", ""},
		{"https://user:password@example.test/owner/repo", ""},
	}

	for _, test := range cases {
		t.Run(test.address, func(t *testing.T) {
			got, err := repositoryURL(test.address)

			if got != test.want || (err != nil) != (test.want == "") {
				t.Fatalf("repository = %q, error = %v; want %q", got, err, test.want)
			}
		})
	}
}
