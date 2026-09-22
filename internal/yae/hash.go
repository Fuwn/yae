package yae

import (
	"context"
	"fmt"
	"strings"
)

func fetchSHA256(context context.Context, address string, unpack bool) (string, error) {
	arguments := []string{"--type", "sha256", address}

	if unpack {
		arguments = append([]string{"--unpack"}, arguments...)
	}

	output, err := command(context, "nix-prefetch-url", arguments...)

	if err != nil {
		return "", err
	}

	hash := strings.TrimSpace(output)

	if !sha256Pattern.MatchString(hash) {
		return "", fmt.Errorf("nix-prefetch-url returned an invalid SHA-256 hash")
	}

	return hash, nil
}

func fetchSRIHash(context context.Context, sha256 string) (string, error) {
	output, err := command(context, "nix", "hash", "convert", "--hash-algo", "sha256", "--from", "nix32", "--to", "sri", sha256)

	if err != nil {
		return "", err
	}

	hash := strings.TrimSpace(output)

	if err := validateSRIHash(hash); err != nil {
		return "", fmt.Errorf("nix hash convert: %w", err)
	}

	return hash, nil
}
