package store

import (
	"os"
	"path/filepath"
)

const (
	appDirName   = "ctrl"
	todosFile    = "todos.json"
	configEnvVar = "CTRL_CONFIG_HOME"
)

func DefaultConfigDir() (string, error) {
	if override := os.Getenv(configEnvVar); override != "" {
		return override, nil
	}

	dir, err := os.UserConfigDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, appDirName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", appDirName), nil
}

func DefaultTodosPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, todosFile), nil
}
