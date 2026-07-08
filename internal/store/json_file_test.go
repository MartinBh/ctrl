package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteIndentedJSONReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "data.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("old content\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	value := struct {
		Name string `json:"name"`
	}{
		Name: "ctrl",
	}
	if err := writeIndentedJSON(path, value); err != nil {
		t.Fatalf("writeIndentedJSON() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	want := "{\n  \"name\": \"ctrl\"\n}\n"
	if string(data) != want {
		t.Fatalf("written JSON = %q, want %q", string(data), want)
	}

	tempFiles := matchingTempFiles(t, filepath.Dir(path), filepath.Base(path))
	if len(tempFiles) != 0 {
		t.Fatalf("left temp files after successful write: %v", tempFiles)
	}
}

func TestWriteFileReplaceCleansTempFileOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	if err := writeFileReplace(path, []byte("{}\n"), filePerm); err == nil {
		t.Fatal("writeFileReplace() error = nil, want rename error")
	}

	tempFiles := matchingTempFiles(t, dir, filepath.Base(path))
	if len(tempFiles) != 0 {
		t.Fatalf("left temp files after failed write: %v", tempFiles)
	}
}

func matchingTempFiles(t *testing.T, dir string, base string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	var matches []string
	prefix := "." + base + "."
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".tmp") {
			matches = append(matches, entry.Name())
		}
	}

	return matches
}
