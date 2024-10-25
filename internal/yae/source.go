package yae

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
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

func (source *Source) Update(sources *Environment, name string, force bool, forcePinned bool) (bool, error) {
	log.Infof("checking %s", name)

	updated := false

	if !sources.Exists(name) {
		log.Warnf("skipped %s: source does not exist", name)

		return updated, nil
	}

	if source.Pinned && !forcePinned {
		log.Infof("skipped %s: source is pinned", name)

		return updated, nil
	}

	if source.Type == "git" {
		log.Debugf("checking %s: remote git tag", name)

		tag, err := source.fetchLatestGitTag()

		if err != nil {
			return updated, err
		}

		if tag != source.Version || force || source.Force {
			if tag != source.Version {
				log.Infof("bumped %s: %s -> %s", name, source.Version, tag)
			}

			if tag != source.Version {
				updated = true
			}

			source.Version = tag

			if strings.Contains(source.URLTemplate, "{version}") {
				source.URL = strings.ReplaceAll(source.URLTemplate, "{version}", source.Version)

				log.Debugf("patched %s: substituted url template", name)
			}
		} else {
			log.Infof("skipped %s: version remains unchanged", name)

			return updated, nil
		}
	}

	log.Debugf("checking %s: sha256", name)

	sha256, err := FetchSHA256(source.URL, source.Unpack)

	if err != nil {
		return updated, err
	}

	if sha256 != source.SHA256 {
		log.Infof("rehashed %s: %s -> %s", name, source.SHA256, sha256)

		source.SHA256 = sha256
		updated = true
	}

	(*sources).Sources[name] = *source

	return updated, nil
}

func (source *Source) fetchLatestGitTag() (string, error) {
	if source.Type == "git" {
		url, err := url.Parse(source.URL)

		if err != nil {
			return "", err
		}

		domain := url.Host
		pathSegments := strings.Split(url.Path, "/")
		repository := url.Scheme + "://" + domain + "/" + pathSegments[1] + "/" + pathSegments[2]
		remotes, err := command("bash", false, "-c", fmt.Sprintf("git ls-remote %s | awk -F'/' '{print $NF}' | sort -V", repository))

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
