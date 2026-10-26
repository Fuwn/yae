package yae

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

const nixBase32Alphabet = "0123456789abcdfghijklmnpqrsvwxyz"

var sha256Pattern = regexp.MustCompile(`^[01][` + nixBase32Alphabet + `]{51}$`)

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

// sriHash converts a Nix base32 SHA-256 hash to its SRI form. The input must
// match sha256Pattern. Nix base32 stores the digest's bits little-endian,
// starting from the final character.
func sriHash(nixBase32 string) string {
	var digest [32]byte

	for position := range len(nixBase32) {
		digit := strings.IndexByte(nixBase32Alphabet, nixBase32[len(nixBase32)-1-position])
		bitOffset := position * 5
		byteIndex, shift := bitOffset/8, bitOffset%8

		digest[byteIndex] |= byte(digit << shift)

		if byteIndex+1 < len(digest) {
			digest[byteIndex+1] |= byte(digit >> (8 - shift))
		}
	}

	return "sha256-" + base64.StdEncoding.EncodeToString(digest[:])
}
