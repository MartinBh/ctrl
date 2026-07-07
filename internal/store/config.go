package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	TutorialSeen    bool   `json:"tutorial_seen"`
	RefreshInterval string `json:"refresh_interval,omitempty"`
	Theme           string `json:"theme,omitempty"`
}

type ConfigStore struct {
	path string
}

func NewConfigStore(path string) *ConfigStore {
	return &ConfigStore{path: path}
}

func (s *ConfigStore) Load() (Config, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	if len(data) == 0 {
		return Config{}, nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (s *ConfigStore) Save(config Config) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}
