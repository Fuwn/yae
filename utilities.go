package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func fetchSHA256(url string, unpack bool) (string, error) {
	arguments := []string{"--type", "sha256", url}

	if unpack {
		arguments = append([]string{"--unpack"}, arguments...)
	}

	output, err := command("nix-prefetch-url", false, arguments...)

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

func lister(items []string) string {
	if len(items) == 0 {
		return ""
	} else if len(items) == 1 {
		return items[0]
	} else if len(items) == 2 {
		return fmt.Sprintf("%s & %s", items[0], items[1])
	}

	return fmt.Sprintf("%s, & %s", strings.Join(items[:len(items)-1], ", "), items[len(items)-1])
}
