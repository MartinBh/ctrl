package todo

import (
	"testing"

	"github.com/martinbhatta/ctrl/internal/store"
)

func TestPanelShowsActionableEmptyState(t *testing.T) {
	panel := NewPanel()

	panel.SetTodos(nil)

	if panel.list.GetItemCount() != 1 {
		t.Fatalf("item count = %d, want 1", panel.list.GetItemCount())
	}
	main, _ := panel.list.GetItemText(0)
	if main != "[gray]No todos yet. Press a to add one." {
		t.Fatalf("empty state = %q", main)
	}
	if _, ok := panel.SelectedTodo(); ok {
		t.Fatal("SelectedTodo() returned a todo for empty state")
	}
}

func TestPanelTracksSelectedTodo(t *testing.T) {
	panel := NewPanel()
	panel.SetTodos([]store.Todo{
		{ID: "one", Title: "First"},
		{ID: "two", Title: "Second"},
	})

	if ok := panel.SelectTodo("two"); !ok {
		t.Fatal("SelectTodo(two) = false, want true")
	}

	todo, ok := panel.SelectedTodo()
	if !ok {
		t.Fatal("SelectedTodo() ok = false, want true")
	}
	if todo.ID != "two" {
		t.Fatalf("SelectedTodo() ID = %q, want %q", todo.ID, "two")
	}
}

func TestPanelPreservesSelectionWhenTodosRefresh(t *testing.T) {
	panel := NewPanel()
	panel.SetTodos([]store.Todo{
		{ID: "one", Title: "First"},
		{ID: "two", Title: "Second"},
	})
	panel.SelectTodo("two")

	panel.SetTodos([]store.Todo{
		{ID: "one", Title: "First edited"},
		{ID: "two", Title: "Second edited"},
		{ID: "three", Title: "Third"},
	})

	todo, ok := panel.SelectedTodo()
	if !ok {
		t.Fatal("SelectedTodo() ok = false, want true")
	}
	if todo.ID != "two" {
		t.Fatalf("SelectedTodo() ID = %q, want %q", todo.ID, "two")
	}
}

func TestPanelMovesSelectionWhenSelectedTodoDisappears(t *testing.T) {
	panel := NewPanel()
	panel.SetTodos([]store.Todo{
		{ID: "one", Title: "First"},
		{ID: "two", Title: "Second"},
	})
	panel.SelectTodo("two")

	panel.SetTodos([]store.Todo{
		{ID: "one", Title: "First"},
	})

	todo, ok := panel.SelectedTodo()
	if !ok {
		t.Fatal("SelectedTodo() ok = false, want true")
	}
	if todo.ID != "one" {
		t.Fatalf("SelectedTodo() ID = %q, want %q", todo.ID, "one")
	}
}

func TestPanelSetErrorClearsSelectedTodo(t *testing.T) {
	panel := NewPanel()
	panel.SetTodos([]store.Todo{{ID: "one", Title: "First"}})

	panel.SetError(assertErr("boom"))

	if _, ok := panel.SelectedTodo(); ok {
		t.Fatal("SelectedTodo() returned a todo after SetError")
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
