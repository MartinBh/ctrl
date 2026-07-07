package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStoreLoadMissingFileReturnsDefaults(t *testing.T) {
	store := NewConfigStore(filepath.Join(t.TempDir(), "config.json"))

	config, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.TutorialSeen {
		t.Fatal("TutorialSeen = true, want false")
	}
}

func TestConfigStoreSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	store := NewConfigStore(path)
	want := Config{
		TutorialSeen:    true,
		RefreshInterval: "5m",
		Theme:           "green-crt",
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatal("saved config should end with a newline")
	}
}
