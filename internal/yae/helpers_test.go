package yae

import (
	"path/filepath"
	"testing"

	"github.com/Fuwn/yae/internal/fixture"
)

func validSource() Source {
	return Source{URL: "https://example.test/file", SHA256: fixture.DigestSHA256, Hash: fixture.DigestSRIHash, Type: "binary", Unpack: true}
}

func fakeGit(t *testing.T, tags string, exitWithFailure bool) (argumentsPath string) {
	t.Helper()

	argumentsPath = filepath.Join(t.TempDir(), "arguments")

	script := "printf '%s\\n' \"$@\" > \"$YAE_TEST_ARGUMENTS\"\nprintf '%s' \"$YAE_TEST_TAGS\"\n"

	if exitWithFailure {
		script += "exit 23\n"
	}

	fixture.FakeCommand(t, "git", script)
	t.Setenv("YAE_TEST_ARGUMENTS", argumentsPath)
	t.Setenv("YAE_TEST_TAGS", tags)

	return argumentsPath
}
