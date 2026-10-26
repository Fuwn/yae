package yae

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Environment struct {
	Schema  string
	Sources map[string]Source
}

func (environment Environment) CheckNewName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}

	if _, exists := environment.Sources[name]; exists {
		return fmt.Errorf("source %q already exists", name)
	}

	return nil
}

func (environment Environment) validate() error {
	for name, source := range environment.Sources {
		if err := validateName(name); err != nil {
			return err
		}

		if err := source.validate(); err != nil {
			return fmt.Errorf("source %q: %w", name, err)
		}
	}

	return nil
}

func (environment *Environment) Save(path string) error {
	if err := environment.validate(); err != nil {
		return err
	}

	contents := make(map[string]any, len(environment.Sources)+1)

	for name, source := range environment.Sources {
		contents[name] = source
	}

	if environment.Schema != "" {
		contents["$schema"] = environment.Schema
	}

	data, err := json.MarshalIndent(contents, "", "  ")

	if err != nil {
		return err
	}

	mode := os.FileMode(0o600)

	if information, err := os.Stat(path); err == nil {
		if !information.Mode().IsRegular() {
			return fmt.Errorf("sources path must be a regular file")
		}

		mode = information.Mode().Perm()
		path, err = filepath.EvalSymlinks(path)

		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	} else if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("sources path is a dangling symbolic link")
	} else if !os.IsNotExist(err) {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".yae-*")

	if err != nil {
		return err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	if err := file.Chmod(mode); err != nil {
		return err
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), path)
}

func (environment *Environment) Load(path string) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	var contents map[string]json.RawMessage

	if err := json.Unmarshal(data, &contents); err != nil {
		return err
	}

	loaded := Environment{Sources: make(map[string]Source, len(contents))}

	for name, data := range contents {
		if name == "$schema" {
			if err := json.Unmarshal(data, &loaded.Schema); err != nil {
				return fmt.Errorf("invalid $schema: %w", err)
			}

			continue
		}

		var source Source

		decoder := json.NewDecoder(bytes.NewReader(data))

		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&source); err != nil {
			return fmt.Errorf("source %q: %w", name, err)
		}

		loaded.Sources[name] = source
	}

	if err := loaded.validate(); err != nil {
		return err
	}

	*environment = loaded

	return nil
}
