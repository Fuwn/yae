package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Sources map[string]Source

type Source struct {
	Url    string `json:"url"`
	SHA256 string `json:"sha256"`
	Unpack bool   `json:"unpack"`
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
