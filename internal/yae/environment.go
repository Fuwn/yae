package yae

import (
	"encoding/json"
	"fmt"
	"os"
)

type Environment struct {
	Schema  string
	Sources map[string]Source
}

func (s *Environment) Add(name string, d Source) error {
	if s.Exists(name) {
		return fmt.Errorf("source already exists")
	}

	(*s).Sources[name] = d

	return nil
}

func (s *Environment) Exists(name string) bool {
	_, ok := (*s).Sources[name]

	return ok
}

func (s *Environment) Drop(url string) {
	delete((*s).Sources, url)
}

func (s *Environment) Save(path string) error {
	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	sourcesData, err := json.Marshal(s.Sources)

	if err != nil {
		return err
	}

	var jsonData map[string]json.RawMessage

	if err := json.Unmarshal(sourcesData, &jsonData); err != nil {
		return err
	}

	if s.Schema != "" {
		jsonData["$schema"] = json.RawMessage(fmt.Sprintf(`"%s"`, s.Schema))
	}

	encoder := json.NewEncoder(file)

	encoder.SetIndent("", "  ")

	return encoder.Encode(jsonData)
}

func (s *Environment) Load(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	var rawData map[string]json.RawMessage

	if err := json.NewDecoder(file).Decode(&rawData); err != nil {
		return err
	}

	if schema, ok := rawData["$schema"]; ok {
		json.Unmarshal(schema, &s.Schema)
	}

	delete(rawData, "$schema")

	if filteredData, err := json.Marshal(rawData); err != nil {
		return err
	} else {
		return json.Unmarshal(filteredData, &s.Sources)
	}
}
