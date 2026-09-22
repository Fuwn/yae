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
	Force         bool   `json:"force,omitempty"`
}

func (source Source) Update(context context.Context, forceHash bool, forcePinned bool) (Source, error) {
	if source.Pinned && !forcePinned {
		return source.repairHash(context)
	}

	if source.Type == "git" {
		tag, err := source.fetchLatestGitTag(context)

		if err != nil {
			return Source{}, err
		}

		if tag == source.Version && !forceHash && !source.Force {
			return source.repairHash(context)
		}

		source.Version = tag
		source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", tag)
	}

	if err := source.RefreshHashes(context); err != nil {
		return Source{}, err
	}

	return source, nil
}

func (source *Source) fetchLatestGitTag(context context.Context) (string, error) {
	if source.Type != "git" {
		return "", fmt.Errorf("source is not a git repository")
	}

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

		if len(fields) != 2 || !strings.HasPrefix(fields[1], "refs/tags/") || strings.HasSuffix(fields[1], "^{}") {
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

	return "", fmt.Errorf("no remote tags match tag_predicate")
}

func repositoryURL(address string) (string, error) {
	parsed, err := url.Parse(address)

	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil {
		return "", fmt.Errorf("git source requires an HTTP(S) URL without credentials")
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	repositoryEnd := len(segments)

	for index, segment := range segments {
		if index >= 2 && (segment == "archive" || segment == "releases" || segment == "-") {
			repositoryEnd = index

			break
		}
	}

	if repositoryEnd < 2 || (repositoryEnd == len(segments) && repositoryEnd != 2) {
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

	hash, err := fetchSRIHash(context, sha256)

	if err != nil {
		return err
	}

	source.SHA256 = sha256
	source.Hash = hash

	return nil
}

func (source Source) repairHash(context context.Context) (Source, error) {
	if source.Hash != "" {
		return source, nil
	}

	hash, err := fetchSRIHash(context, source.SHA256)

	if err != nil {
		return Source{}, err
	}

	source.Hash = hash

	return source, nil
}
