package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTodoStoreLoadMissingFileReturnsEmptyList(t *testing.T) {
	store := NewTodoStore(filepath.Join(t.TempDir(), "todos.json"))

	todos, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(todos) != 0 {
		t.Fatalf("Load() returned %d todos, want 0", len(todos))
	}
}

func TestTodoStoreSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "todos.json")
	store := NewTodoStore(path)
	want := []Todo{
		{
			ID:        "todo-1",
			Title:     "Write tests",
			Done:      true,
			CreatedAt: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		},
	}

	if err := store.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatal("saved todos should end with a newline")
	}
}

func TestTodoStoreAddTrimsAndPersistsTodo(t *testing.T) {
	store := NewTodoStore(filepath.Join(t.TempDir(), "todos.json"))

	todos, todo, err := store.Add("  Ship feature  ")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if todo.ID == "" {
		t.Fatal("Add() returned empty todo ID")
	}
	if todo.CreatedAt.IsZero() {
		t.Fatal("Add() returned zero CreatedAt")
	}
	if todo.Title != "Ship feature" {
		t.Fatalf("Add() title = %q, want %q", todo.Title, "Ship feature")
	}
	if len(todos) != 1 || todos[0] != todo {
		t.Fatalf("Add() todos = %#v, want appended todo %#v", todos, todo)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 1 || loaded[0] != todo {
		t.Fatalf("Load() = %#v, want %#v", loaded, []Todo{todo})
	}
}

func TestTodoStoreAddRejectsBlankTitle(t *testing.T) {
	store := NewTodoStore(filepath.Join(t.TempDir(), "todos.json"))

	if _, _, err := store.Add(" \t "); !errors.Is(err, ErrEmptyTodoTitle) {
		t.Fatalf("Add() error = %v, want ErrEmptyTodoTitle", err)
	}
}

func TestTodoStoreUpdateToggleAndDelete(t *testing.T) {
	store := NewTodoStore(filepath.Join(t.TempDir(), "todos.json"))
	createdAt := time.Date(2026, 7, 8, 11, 0, 0, 0, time.UTC)
	todos := []Todo{
		{ID: "todo-1", Title: "First", CreatedAt: createdAt},
		{ID: "todo-2", Title: "Second", CreatedAt: createdAt},
	}
	if err := store.Save(todos); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	updated, err := store.UpdateTitle("todo-1", "  First edited  ")
	if err != nil {
		t.Fatalf("UpdateTitle() error = %v", err)
	}
	if updated[0].Title != "First edited" {
		t.Fatalf("UpdateTitle() title = %q, want %q", updated[0].Title, "First edited")
	}

	toggled, err := store.Toggle("todo-2")
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}
	if !toggled[1].Done {
		t.Fatal("Toggle() left todo-2 incomplete, want complete")
	}

	deleted, err := store.Delete("todo-1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(deleted) != 1 || deleted[0].ID != "todo-2" {
		t.Fatalf("Delete() = %#v, want only todo-2", deleted)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "todo-2" || !loaded[0].Done {
		t.Fatalf("Load() = %#v, want completed todo-2", loaded)
	}
}

func TestTodoStoreMutationsReturnNotFound(t *testing.T) {
	store := NewTodoStore(filepath.Join(t.TempDir(), "todos.json"))

	if _, err := store.UpdateTitle("missing", "Title"); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("UpdateTitle() error = %v, want ErrTodoNotFound", err)
	}
	if _, err := store.Toggle("missing"); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("Toggle() error = %v, want ErrTodoNotFound", err)
	}
	if _, err := store.Delete("missing"); !errors.Is(err, ErrTodoNotFound) {
		t.Fatalf("Delete() error = %v, want ErrTodoNotFound", err)
	}
}

func TestTodoStoreUpdateRejectsBlankTitle(t *testing.T) {
	store := NewTodoStore(filepath.Join(t.TempDir(), "todos.json"))

	if _, err := store.UpdateTitle("todo-1", " "); !errors.Is(err, ErrEmptyTodoTitle) {
		t.Fatalf("UpdateTitle() error = %v, want ErrEmptyTodoTitle", err)
	}
}
