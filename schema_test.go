package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestSchema(t *testing.T) {
	schema, err := jsonschema.Compile("yae.schema.json")

	if err != nil {
		t.Fatal(err)
	}

	read := func(path string) map[string]any {
		t.Helper()

		contents, err := os.ReadFile(path)

		if err != nil {
			t.Fatal(err)
		}

		var environment map[string]any

		if err := json.Unmarshal(contents, &environment); err != nil {
			t.Fatal(err)
		}

		return environment
	}
	paths, err := filepath.Glob("examples/*/yae.json")

	if err != nil || len(paths) == 0 {
		t.Fatalf("example discovery: %v", err)
	}

	for _, path := range paths {
		if err := schema.Validate(read(path)); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
	}

	for _, invalid := range []any{nil, map[string]any{"$schema": nil}, map[string]any{"source": nil}, map[string]any{"source": map[string]any{}}} {
		if schema.Validate(invalid) == nil {
			t.Fatalf("schema accepted invalid environment: %#v", invalid)
		}
	}

	for field, value := range map[string]any{
		"type": "unsupported", "sha256": "invalid", "hash": "sha256-invalid",
		"unpack": nil, "future_option": true,
	} {
		invalid := read("examples/nixpkgs/yae.json")

		invalid["nixpkgs"].(map[string]any)[field] = value

		if schema.Validate(invalid) == nil {
			t.Fatalf("schema accepted invalid %s", field)
		}
	}

	legacy := read("examples/nixpkgs/yae.json")
	source := legacy["nixpkgs"].(map[string]any)

	source["force"] = true
	source["trim_tag_prefix"] = "v"
	source["tag_predicate"] = "^v"

	if err := schema.Validate(legacy); err != nil {
		t.Fatal(err)
	}
}
