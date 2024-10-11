package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Sources map[string]Source

type Source struct {
	URI           string `json:"url"`
	SHA256        string `json:"sha256"`
	Unpack        bool   `json:"unpack"`
	Type          string `json:"type"`
	Version       string `json:"version,omitempty"`
	URITemplate   string `json:"uri_template,omitempty"`
	TagPredicate  string `json:"tag_predicate,omitempty"`
	TrimTagPrefix string `json:"trim_tag_prefix,omitempty"`
	Pinned        bool   `json:"pinned,omitempty"`
}

func (s *Sources) EnsureLoaded() error {
	return nil
}

func (s *Sources) Add(name string, d Source) error {
	if s.Exists(name) {
		return fmt.Errorf("source already exists")
	}

	(*s)[name] = d

	return nil
}

func (s *Sources) Exists(name string) bool {
	_, ok := (*s)[name]

	return ok
}

func (s *Sources) Drop(url string) {
	delete((*s), url)
}

func (s *Sources) Save(path string) error {
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)

	encoder.SetIndent("", "  ")

	return encoder.Encode(s)
}

func (s *Sources) Load(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(s)
}
