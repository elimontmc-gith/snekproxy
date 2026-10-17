package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Enabled bool `yaml:"enabled"`
	Verbose bool `yaml:"verbose"`
}

func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, errors.New("configuration path must not be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read configuration %q: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return Config{}, fmt.Errorf("configuration %q is empty", path)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration %q: %w", path, err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("decode configuration %q: multiple YAML documents are not supported", path)
		}
		return Config{}, fmt.Errorf("decode configuration %q: %w", path, err)
	}

	return cfg, nil
}
