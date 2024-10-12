package main

import (
	"os"
	"os/exec"
	"strings"
)

func fetchSHA256(url string, unpack bool, show bool) (string, error) {
	arguments := []string{"--type", "sha256", url}

	if unpack {
		arguments = append([]string{"--unpack"}, arguments...)
	}

	output, err := command("nix-prefetch-url", show, arguments...)

	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")

	return strings.Trim(lines[len(lines)-2], "\n"), nil
}

func command(name string, show bool, args ...string) (string, error) {
	executable, err := exec.LookPath(name)
	out := []byte{}

	if show {
		cmd := exec.Command(executable, args...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err = cmd.Output()
	} else {
		cmd := exec.Command(executable, args...)
		out, err = cmd.Output()
	}

	return string(out), err
}
