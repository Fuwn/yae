package yae

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSHA256 = "1wn29537l343lb0id0byk0699fj0k07m1n2d7jx2n0ssax55vhwy"
const testSRIHash = "sha256-nsNdSldaAyu6PE3YUA+YQLqUDJh+gRbBooMMekZJwvI="

func validSource() Source {
	return Source{URL: "https://example.test/file", SHA256: testSHA256, Hash: testSRIHash, Type: "binary", Unpack: true}
}

func fakeNix(t *testing.T) {
	t.Helper()

	for name, output := range map[string]string{"nix-prefetch-url": testSHA256, "nix": testSRIHash} {
		fakeCommand(t, name, "printf '%s\\n' '"+output+"'")
	}
}

func TestInvalidStoredSources(t *testing.T) {
	cases := map[string]string{
		"null environment": `null`,
		"null schema":      `{"$schema":null}`,
		"reserved name":    `{"$schema":{}}`,
		"null source":      `{"sample":null}`,
	}
	changes := map[string]map[string]any{
		"unknown field":    {"future_option": true},
		"unknown type":     {"type": "other"},
		"bad sha256":       {"sha256": "invalid"},
		"bad sri":          {"hash": "sha256-invalid"},
		"null bool":        {"unpack": nil},
		"null optional":    {"pinned": nil},
		"bad url":          {"url": "relative/path"},
		"credentials":      {"url": "https://user:password@example.test/file"},
		"missing template": {"type": "git", "version": "v1"},
		"conflicting pin":  {"pinned": true, "force": true},
		"mismatched url":   {"version": "v1", "url_template": "https://example.test/{version}"},
	}

	for name, fields := range changes {
		data, err := json.Marshal(validSource())

		if err != nil {
			t.Fatal(err)
		}

		var source map[string]any

		if err := json.Unmarshal(data, &source); err != nil {
			t.Fatal(err)
		}

		for field, value := range fields {
			source[field] = value
		}

		data, err = json.Marshal(map[string]any{"sample": source})

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

func TestAddInitializesMapAndRejectsReservedName(t *testing.T) {
	environment := Environment{}

	if err := environment.Add("$schema", validSource()); err == nil {
		t.Fatal("reserved name accepted")
	}

	if err := environment.Add("sample", validSource()); err != nil {
		t.Fatal(err)
	}

	if err := environment.Add("sample", validSource()); err == nil {
		t.Fatal("duplicate name accepted")
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

func TestRefreshPopulatesBothHashes(t *testing.T) {
	fakeNix(t)

	source := Source{URL: "https://example.test/file", Type: "binary"}

	if err := source.RefreshHashes(context.Background()); err != nil {
		t.Fatal(err)
	}

	if source.SHA256 != testSHA256 || source.Hash != testSRIHash {
		t.Fatalf("incomplete hashes: %#v", source)
	}
}

func TestUpdateRepairsLegacyHash(t *testing.T) {
	cases := []struct {
		name   string
		pinned bool
	}{
		{name: "unchanged version"},
		{name: "pinned", pinned: true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fakeNix(t)
			fakeGit(t, "abc\trefs/tags/v1\n", false)

			source := Source{URL: "https://example.test/owner/repo/archive/v1", URLTemplate: "https://example.test/owner/repo/archive/{version}", Version: "v1", Type: "git", SHA256: testSHA256, Pinned: test.pinned}
			updated, err := source.Update(context.Background(), false, false)

			if err != nil || updated.Hash != testSRIHash || updated.SHA256 != testSHA256 || updated.Version != "v1" {
				t.Fatalf("legacy repair failed: %#v, %v, %v", source, updated, err)
			}
		})
	}
}

func TestMissingRequiredField(t *testing.T) {
	data, err := json.Marshal(validSource())

	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "sources.json")
	contents := `{"sample":` + strings.Replace(string(data), `"unpack":true,`, "", 1) + `}`

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	var environment Environment

	if err := environment.Load(path); err == nil {
		t.Fatal("missing unpack accepted")
	}
}

func TestLegacyBinaryOptionsRoundTrip(t *testing.T) {
	for _, option := range []string{"force", "trim_tag_prefix", "tag_predicate"} {
		t.Run(option, func(t *testing.T) {
			source := validSource()

			switch option {
			case "force":
				source.Force = true
			case "trim_tag_prefix":
				source.TrimTagPrefix = "v"
			case "tag_predicate":
				source.TagPredicate = "^v"
			}

			if err := source.Validate(); err == nil {
				t.Fatal("new binary source accepted irrelevant options")
			}

			path := filepath.Join(t.TempDir(), "sources.json")
			original := Environment{Sources: map[string]Source{"legacy": source}}

			if err := original.Save(path); err != nil {
				t.Fatal(err)
			}

			var loaded Environment

			if err := loaded.Load(path); err != nil || loaded.Sources["legacy"] != source {
				t.Fatalf("legacy source changed: %#v, %v", loaded, err)
			}

			loaded.Drop("legacy")

			if err := loaded.Save(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}
