// Package fixture provides fake external commands and a known hash pair for
// tests.
package fixture

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// SHA256 and SRIHash are the Nix base32 and SRI forms of one real digest.
const (
	SHA256  = "1wn29537l343lb0id0byk0699fj0k07m1n2d7jx2n0ssax55vhwy"
	SRIHash = "sha256-nsNdSldaAyu6PE3YUA+YQLqUDJh+gRbBooMMekZJwvI="
)

// Command places a shell script called name at the front of PATH for the rest
// of the test.
func Command(t testing.TB, name string, body string) {
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
