package yae

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var sha256Pattern = regexp.MustCompile(`^[01][0123456789abcdfghijklmnpqrsvwxyz]{51}$`)

func (source Source) Validate() error {
	if source.Type != "git" && (source.TagPredicate != "" || source.TrimTagPrefix != "" || source.Force) {
		return fmt.Errorf("tag_predicate, trim_tag_prefix, and force require a git source")
	}

	return source.validateConfiguration()
}

func (source Source) validateConfiguration() error {
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

	if source.Pinned && source.Force {
		return fmt.Errorf("source cannot be both pinned and forced")
	}

	if source.Version != "" || source.URLTemplate != "" {
		if source.Version == "" || !strings.Contains(source.URLTemplate, "{version}") {
			return fmt.Errorf("version and url_template containing {version} must be supplied together")
		}

		if source.URL != strings.ReplaceAll(source.URLTemplate, "{version}", source.Version) {
			return fmt.Errorf("url does not match url_template and version")
		}
	}

	if source.Type == "git" {
		if source.Version == "" {
			return fmt.Errorf("git source requires version and url_template")
		}

		if _, err := repositoryURL(source.URL); err != nil {
			return err
		}

		if _, err := regexp.Compile(source.TagPredicate); err != nil {
			return fmt.Errorf("invalid tag_predicate: %w", err)
		}
	}

	return nil
}

func (source Source) validateStored() error {
	if err := source.validateConfiguration(); err != nil {
		return err
	}

	if !sha256Pattern.MatchString(source.SHA256) {
		return fmt.Errorf("sha256 must be a 52-character Nix base32 SHA-256 hash")
	}

	if source.Hash == "" {
		return nil
	}

	if !strings.HasPrefix(source.Hash, "sha256-") {
		return fmt.Errorf("hash must be a SHA-256 SRI hash")
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(source.Hash, "sha256-"))

	if err != nil || len(decoded) != 32 || "sha256-"+base64.StdEncoding.EncodeToString(decoded) != source.Hash {
		return fmt.Errorf("hash must be a SHA-256 SRI hash")
	}

	return nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" || name == "$schema" {
		return fmt.Errorf("source name must be non-empty and cannot be $schema")
	}

	return nil
}
