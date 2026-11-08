package yae

import (
	"context"
	"strings"
	"testing"

	"github.com/Fuwn/yae/internal/fixture"
)

func TestHashOutputValidation(t *testing.T) {
	for _, output := range []string{"", "invalid", fixture.DigestSHA256 + "\nextra", fixture.DigestSHA256, fixture.DigestSHA256 + "\n"} {
		t.Run(output, func(t *testing.T) {
			fixture.FakeCommand(t, "nix-prefetch-url", "printf '%s' \"$YAE_TEST_HASH\"")
			t.Setenv("YAE_TEST_HASH", output)

			hash, err := fetchSHA256(context.Background(), "https://example.test/file", false)
			valid := strings.TrimSpace(output) == fixture.DigestSHA256

			if (err == nil) != valid || (valid && hash != fixture.DigestSHA256) {
				t.Fatalf("hash = %q, error = %v", hash, err)
			}
		})
	}
}

func TestSRIHash(t *testing.T) {
	cases := map[string]string{
		fixture.DigestSHA256: fixture.DigestSRIHash,
		"0000000000000000000000000000000000000000000000000000": "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"0zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz": "sha256-/////////////////////////////////////////38=",
		"1000000000000000000000000000000000000000000000000000": "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIA=",
	}

	for nixBase32, want := range cases {
		if got := sriHash(nixBase32); got != want {
			t.Errorf("sriHash(%s) = %s; want %s", nixBase32, got, want)
		}
	}
}
