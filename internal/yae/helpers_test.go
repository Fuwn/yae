package yae

import (
	"path/filepath"
	"testing"

	"github.com/Fuwn/yae/internal/fixture"
)

func validSource() Source {
	return Source{URL: "https://example.test/file", SHA256: fixture.SHA256, Hash: fixture.SRIHash, Type: "binary", Unpack: true}
}

// fakeGit makes git print tags, optionally fail, and record its arguments in
// the returned file.
func fakeGit(t *testing.T, tags string, failure bool) string {
	t.Helper()

	arguments := filepath.Join(t.TempDir(), "arguments")
	script := "printf '%s\\n' \"$@\" > \"$YAE_TEST_ARGUMENTS\"\nprintf '%s' \"$YAE_TEST_TAGS\"\n"

	if failure {
		script += "exit 23\n"
	}

	fixture.Command(t, "git", script)
	t.Setenv("YAE_TEST_ARGUMENTS", arguments)
	t.Setenv("YAE_TEST_TAGS", tags)

	return arguments
}
