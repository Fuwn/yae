package yae

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func fakeGit(t *testing.T, output string, failure bool) string {
	t.Helper()

	arguments := filepath.Join(t.TempDir(), "arguments")
	script := "printf '%s\\n' \"$@\" > \"$YAE_TEST_ARGUMENTS\"\nprintf '%s' \"$YAE_TEST_TAGS\"\n"

	if failure {
		script += "exit 23\n"
	}

	fakeCommand(t, "git", script)
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
			version, err := source.fetchLatestGitTag(context.Background())

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

	if _, err := source.fetchLatestGitTag(context.Background()); err != nil {
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

func TestGitVersionOrderingWithLocalRepository(t *testing.T) {
	repository := t.TempDir()
	runGit := func(input string, arguments ...string) string {
		t.Helper()

		process := exec.Command("git", append([]string{"-C", repository}, arguments...)...)

		process.Stdin = strings.NewReader(input)

		output, err := process.CombinedOutput()

		if err != nil {
			t.Fatalf("git %v: %s, %v", arguments, output, err)
		}

		return strings.TrimSpace(string(output))
	}

	runGit("", "init", "--bare")

	tree := runGit("", "mktree")
	commit := runGit("", "-c", "user.name=Yae Test", "-c", "user.email=yae@example.test", "commit-tree", tree, "-m", "fixture")

	runGit("", "update-ref", "refs/tags/v2", commit)
	runGit("", "update-ref", "refs/tags/v10", commit)
	runGit("", "update-ref", "refs/heads/zzbranch", commit)
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "url.file://"+repository+".insteadOf")
	t.Setenv("GIT_CONFIG_VALUE_0", "https://example.test/owner/repo")

	source := Source{Type: "git", URL: "https://example.test/owner/repo/archive/v1.tar.gz"}
	version, err := source.fetchLatestGitTag(context.Background())

	if err != nil || version != "v10" {
		t.Fatalf("latest tag = %q, error = %v", version, err)
	}
}

func TestUpdatePolicies(t *testing.T) {
	cases := []struct {
		name        string
		kind        string
		version     string
		pinned      bool
		persistent  bool
		forceHash   bool
		forcePinned bool
		fetch       bool
		wantVersion string
	}{
		{name: "floating URL", kind: "binary", fetch: true},
		{name: "pinned URL", kind: "binary", pinned: true},
		{name: "forced pinned URL", kind: "binary", pinned: true, forcePinned: true, fetch: true},
		{name: "unchanged tag", kind: "git", version: "v2", wantVersion: "v2"},
		{name: "new tag", kind: "git", version: "v1", fetch: true, wantVersion: "v2"},
		{name: "persistent rehash", kind: "git", version: "v2", persistent: true, fetch: true, wantVersion: "v2"},
		{name: "requested rehash", kind: "git", version: "v2", forceHash: true, fetch: true, wantVersion: "v2"},
		{name: "pin overrides rehash", kind: "git", version: "v1", pinned: true, forceHash: true, wantVersion: "v1"},
		{name: "override pin", kind: "git", version: "v1", pinned: true, forcePinned: true, fetch: true, wantVersion: "v2"},
		{name: "pin override alone skips unchanged tag", kind: "git", version: "v2", pinned: true, forcePinned: true, wantVersion: "v2"},
		{name: "both overrides", kind: "git", version: "v2", pinned: true, forcePinned: true, forceHash: true, fetch: true, wantVersion: "v2"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fakeGit(t, "abc\trefs/tags/v2\n", false)
			fakeNix(t)

			marker := filepath.Join(t.TempDir(), "prefetched")

			t.Setenv("YAE_TEST_PREFETCH", marker)
			fakeCommand(t, "nix-prefetch-url", "printf fetched > \"$YAE_TEST_PREFETCH\"\nprintf '%s\\n' '"+testSHA256+"'")

			source := validSource()

			source.Type = test.kind
			source.Pinned = test.pinned
			source.Force = test.persistent

			if test.kind == "git" {
				source.Version = test.version
				source.URLTemplate = "https://example.test/owner/repo/archive/{version}"
				source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", test.version)
			}

			updated, err := source.Update(context.Background(), test.forceHash, test.forcePinned)

			if err != nil || updated.Version != test.wantVersion {
				t.Fatalf("updated = %#v, error = %v", updated, err)
			}

			_, err = os.Stat(marker)

			if (err == nil) != test.fetch {
				t.Fatalf("fetch occurred = %v; want %v", err == nil, test.fetch)
			}
		})
	}
}
