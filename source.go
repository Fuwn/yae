package main

import (
	"fmt"
	"strings"
)

type Source struct {
	URL           string `json:"url"`
	SHA256        string `json:"sha256"`
	Unpack        bool   `json:"unpack"`
	Type          string `json:"type"`
	Version       string `json:"version,omitempty"`
	URLTemplate   string `json:"url_template,omitempty"`
	TagPredicate  string `json:"tag_predicate,omitempty"`
	TrimTagPrefix string `json:"trim_tag_prefix,omitempty"`
	Pinned        bool   `json:"pinned,omitempty"`
	Force         bool   `json:"force,omitempty"`
}

func (source *Source) Update(sources *Sources, name string, show bool, force bool, forcePinned bool) (bool, error) {
	updated := false

	if !sources.Exists(name) {
		return updated, fmt.Errorf("source does not exist")
	}

	if source.Pinned && !forcePinned {
		if show {
			fmt.Println("skipped update for", name, "because it is pinned")
		}

		return updated, nil
	}

	if source.Type == "git" {
		tag, err := source.fetchLatestGitTag(show)

		if err != nil {
			return updated, err
		}

		if tag != source.Version || force || source.Force {
			if show {
				fmt.Println("updated version for", name, "from", source.Version, "to", tag)
			}

			if tag != source.Version {
				updated = true
			}

			source.Version = tag

			if strings.Contains(source.URLTemplate, "{version}") {
				source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", source.Version)
			}
		} else {
			if show {
				fmt.Println("skipped update for", name, "because the version is unchanged")
			}

			return updated, nil
		}
	}

	sha256, err := fetchSHA256(source.URL, source.Unpack, show)

	if err != nil {
		return updated, err
	}

	if sha256 != source.SHA256 {
		if show {
			fmt.Println("updated hash for", name, "from", source.SHA256, "to", sha256)
		}

		source.SHA256 = sha256
		updated = true
	}

	(*sources)[name] = *source

	return updated, nil
}

func (source *Source) fetchLatestGitTag(show bool) (string, error) {
	if source.Type == "git" {
		repository := "https://github.com/" + strings.Split(source.URL, "/")[3] + "/" + strings.Split(source.URL, "/")[4]
		remotes, err := command("bash", show, "-c", fmt.Sprintf("git ls-remote %s | awk -F'/' '{print $NF}' | sort -V", repository))

		if err != nil {
			return "", err
		}

		refs := strings.Split(remotes, "\n")
		var latest string

		if source.TagPredicate == "" {
			latest = refs[len(refs)-2]
		} else {
			for i := len(refs) - 2; i >= 0; i-- {
				if strings.Contains(refs[i], source.TagPredicate) {
					latest = refs[i]

					break
				}
			}
		}

		if source.TrimTagPrefix != "" {
			latest = strings.TrimPrefix(latest, source.TrimTagPrefix)
		}

		return latest, nil
	}

	return "", fmt.Errorf("source is not a git repository")
}
