package yae

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Fuwn/yae/internal/fixture"
)

func TestInvalidStoredSources(t *testing.T) {
	cases := map[string]string{
		"reserved name":   `{"$schema":{}}`,
		"null source":     `{"sample":null}`,
		"invalid schema":  `{"$schema":false}`,
		"trailing object": `{} {}`,
		"invalid source":  `{"sample":false}`,
	}
	changes := map[string]map[string]any{
		"unknown field":      {"future_option": true},
		"removed force":      {"force": true},
		"unknown type":       {"type": "other"},
		"bad sha256":         {"sha256": "invalid"},
		"bad sri":            {"hash": "sha256-invalid"},
		"missing sri":        {"hash": ""},
		"mismatched sri":     {"hash": "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
		"bad url":            {"url": "relative/path"},
		"credentials":        {"url": "https://user:password@example.test/file"},
		"missing template":   {"type": "git", "version": "v1"},
		"mismatched url":     {"version": "v1", "url_template": "https://example.test/{version}"},
		"binary tag options": {"tag_predicate": "^v"},
		"binary tag prefix":  {"trim_tag_prefix": "v"},
	}

	for name, fields := range changes {
		source := map[string]any{
			"url": "https://example.test/file", "type": "binary", "unpack": true,
			"sha256": fixture.SHA256, "hash": fixture.SRIHash,
		}

		for field, value := range fields {
			source[field] = value
		}

		data, err := json.Marshal(map[string]any{"sample": source})

		if err != nil {
			t.Fatal(err)
		}

		cases[name] = string(data)
	}

	for name, contents := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sources.json")
			environment := Environment{Schema: "unchanged"}

			if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}

			if err := environment.Load(path); err == nil || environment.Schema != "unchanged" {
				t.Fatalf("invalid source accepted: %#v, %v", environment, err)
			}
		})
	}
}

func TestUnfilledVersionPlaceholder(t *testing.T) {
	source := validSource()

	source.URL = "https://example.test/{version}/file"

	if err := source.ValidateConfiguration(); err == nil {
		t.Fatal("url with an unfilled {version} accepted")
	}
}

func TestValidationFailurePreservesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sources.json")
	environment := Environment{Sources: map[string]Source{"invalid": {}}}

	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := environment.Save(path); err == nil {
		t.Fatal("invalid source was saved")
	}

	data, err := os.ReadFile(path)

	if err != nil || string(data) != "original" {
		t.Fatalf("file was changed: %q, %v", data, err)
	}
}
