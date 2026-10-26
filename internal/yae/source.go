package yae

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type Source struct {
	URL           string `json:"url"`
	SHA256        string `json:"sha256"`
	Hash          string `json:"hash"`
	Unpack        bool   `json:"unpack"`
	Type          string `json:"type"`
	Version       string `json:"version,omitempty"`
	URLTemplate   string `json:"url_template,omitempty"`
	TagPredicate  string `json:"tag_predicate,omitempty"`
	TrimTagPrefix string `json:"trim_tag_prefix,omitempty"`
	Pinned        bool   `json:"pinned,omitempty"`
}

// Update returns the source with its latest version and hashes. Pinned sources
// are returned unchanged, and a git source whose latest tag is its current
// version is only refetched when rehash is set.
func (source Source) Update(context context.Context, rehash bool) (Source, error) {
	if source.Pinned {
		return source, nil
	}

	if source.Type == "git" {
		tag, err := source.fetchLatestGitTag(context)

		if err != nil {
			return Source{}, err
		}

		if tag == source.Version && !rehash {
			return source, nil
		}

		source.Version = tag
		source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", tag)
	}

	if err := source.RefreshHashes(context); err != nil {
		return Source{}, err
	}

	return source, nil
}

func (source Source) fetchLatestGitTag(context context.Context) (string, error) {
	repository, err := repositoryURL(source.URL)

	if err != nil {
		return "", err
	}

	pattern, err := regexp.Compile(source.TagPredicate)

	if err != nil {
		return "", fmt.Errorf("invalid tag_predicate: %w", err)
	}

	output, err := command(context, "git", "ls-remote", "--tags", "--refs", "--sort=-version:refname", "--", repository)

	if err != nil {
		return "", fmt.Errorf("list remote tags: %w", err)
	}

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := strings.Fields(line)

		if len(fields) != 2 {
			continue
		}

		tag := strings.TrimPrefix(fields[1], "refs/tags/")

		if !pattern.MatchString(tag) {
			continue
		}

		version := strings.TrimPrefix(tag, source.TrimTagPrefix)

		if version == "" {
			return "", fmt.Errorf("tag is empty after trimming its prefix")
		}

		return version, nil
	}

	return "", fmt.Errorf("no remote tags match tag_predicate; use a binary source to follow a branch")
}

func repositoryURL(address string) (string, error) {
	parsed, err := url.Parse(address)

	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil {
		return "", fmt.Errorf("git source requires an HTTP(S) URL without credentials")
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	repositoryEnd := len(segments)

	for index := 2; index < len(segments); index++ {
		if segments[index] == "archive" || segments[index] == "releases" || segments[index] == "-" {
			repositoryEnd = index

			break
		}
	}

	hasDownloadMarker := repositoryEnd < len(segments)
	isRepositoryPath := len(segments) == 2

	if !hasDownloadMarker && !isRepositoryPath {
		return "", fmt.Errorf("cannot infer repository: use an archive or release download URL")
	}

	for _, segment := range segments[:repositoryEnd] {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid repository path")
		}
	}

	parsed.Path = "/" + strings.Join(segments[:repositoryEnd], "/")
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

func (source *Source) RefreshHashes(context context.Context) error {
	sha256, err := fetchSHA256(context, source.URL, source.Unpack)

	if err != nil {
		return err
	}

	source.SHA256 = sha256
	source.Hash = sriHash(sha256)

	return nil
}
