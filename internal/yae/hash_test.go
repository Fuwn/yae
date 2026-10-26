package yae

import (
	"context"
	"strings"
	"testing"
)

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

func TestSRIHash(t *testing.T) {
	cases := map[string]string{
		testSHA256: testSRIHash,
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
