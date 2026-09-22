package yae

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEnvironmentRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.json")
	original := Environment{Schema: "a\"b\\c", Sources: map[string]Source{}}

	if err := original.Save(path); err != nil {
		t.Fatal(err)
	}

	var loaded Environment

	if err := loaded.Load(path); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(original, loaded) {
		t.Fatalf("loaded %#v; want %#v", loaded, original)
	}
}

func TestSavePreservesSymlinkAndPermissions(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target.json")
	link := filepath.Join(directory, "link.json")
	environment := Environment{Sources: map[string]Source{}}

	if err := os.WriteFile(target, []byte("old contents"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink("target.json", link); err != nil {
		t.Fatal(err)
	}

	if err := environment.Save(link); err != nil {
		t.Fatal(err)
	}

	information, err := os.Lstat(link)

	if err != nil || information.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link was replaced: %v", err)
	}

	information, err = os.Stat(target)

	if err != nil || information.Mode().Perm() != 0o600 {
		t.Fatalf("permissions changed: %v", err)
	}

	data, err := os.ReadFile(target)

	if err != nil || string(data) != "{}\n" {
		t.Fatalf("target = %q, error = %v", data, err)
	}

	entries, err := os.ReadDir(directory)

	if err != nil || len(entries) != 2 {
		t.Fatalf("temporary files remain: %v, %v", entries, err)
	}
}

func TestFailedSavePreservesDestination(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "sources.json")
	environment := Environment{}

	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(path, "keep")

	if err := os.WriteFile(marker, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := environment.Save(path); err == nil {
		t.Fatal("expected save to reject a directory")
	}

	data, err := os.ReadFile(marker)

	if err != nil || string(data) != "unchanged" {
		t.Fatalf("destination changed: %q, %v", data, err)
	}
}

func TestNewEnvironmentIsPrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.json")
	environment := Environment{}

	if err := environment.Save(path); err != nil {
		t.Fatal(err)
	}

	information, err := os.Stat(path)

	if err != nil || information.Mode().Perm() != 0o600 {
		t.Fatalf("new environment permissions are not private: %v", err)
	}
}
