package yae

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidateConfiguration checks everything about a source except its hashes, so
// a new source can be rejected before anything is downloaded.
func (source Source) ValidateConfiguration() error {
	if source.Type != "binary" && source.Type != "git" {
		return fmt.Errorf("type must be 'binary' or 'git'")
	}

	address, err := url.Parse(source.URL)

	if err != nil || address.User != nil {
		return fmt.Errorf("url must be valid and must not contain credentials")
	}

	if address.Scheme == "file" {
		if address.Path == "" || !strings.HasPrefix(address.Path, "/") || (address.Host != "" && address.Host != "localhost") {
			return fmt.Errorf("file url must identify a local absolute path")
		}
	} else if (address.Scheme != "https" && address.Scheme != "http") || address.Hostname() == "" {
		return fmt.Errorf("url must use HTTP, HTTPS, or a local file")
	}

	if source.Version != "" || source.URLTemplate != "" {
		if source.Version == "" || !strings.Contains(source.URLTemplate, "{version}") {
			return fmt.Errorf("version and url_template containing {version} must be supplied together")
		}

		if source.URL != strings.ReplaceAll(source.URLTemplate, "{version}", source.Version) {
			return fmt.Errorf("url does not match url_template and version")
		}
	}

	if source.Type == "binary" {
		if source.TagPredicate != "" || source.TrimTagPrefix != "" {
			return fmt.Errorf("tag_predicate and trim_tag_prefix require a git source")
		}

		return nil
	}

	if source.Version == "" {
		return fmt.Errorf("git source requires version and url_template")
	}

	if _, err := repositoryURL(source.URL); err != nil {
		return err
	}

	if _, err := regexp.Compile(source.TagPredicate); err != nil {
		return fmt.Errorf("invalid tag_predicate: %w", err)
	}

	return nil
}

func (source Source) validate() error {
	if err := source.ValidateConfiguration(); err != nil {
		return err
	}

	if !sha256Pattern.MatchString(source.SHA256) {
		return fmt.Errorf("sha256 must be a 52-character Nix base32 SHA-256 hash")
	}

	if source.Hash != sriHash(source.SHA256) {
		return fmt.Errorf("hash must be the SRI form of sha256")
	}

	return nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" || name == "$schema" {
		return fmt.Errorf("source name must be non-empty and cannot be $schema")
	}

	return nil
}
