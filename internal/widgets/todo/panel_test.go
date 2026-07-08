package todo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

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

func TestPanelShowsLoadingState(t *testing.T) {
	panel := NewPanel()
	panel.SetTodos([]store.Todo{{ID: "one", Title: "First"}})

	panel.SetLoading()

	main, _ := panel.list.GetItemText(0)
	if main != "[gray]Loading todos..." {
		t.Fatalf("loading state = %q", main)
	}
	if _, ok := panel.SelectedTodo(); ok {
		t.Fatal("SelectedTodo() returned a todo for loading state")
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

func TestPanelEscapesTodoTitleTags(t *testing.T) {
	title := "[red]fix [blue]bug[-]"
	panel := NewPanel()
	panel.SetTodos([]store.Todo{{ID: "one", Title: title}})

	main, _ := panel.list.GetItemText(0)
	wantMain := "[ ] [red[]fix [blue[]bug[-[]"
	if main != wantMain {
		t.Fatalf("main text = %q, want %q", main, wantMain)
	}

	rendered := drawPanel(t, panel, 48, 5)
	if !strings.Contains(rendered, "[red]fix [blue]bug[-]") {
		t.Fatalf("rendered panel missing literal title in:\n%s", rendered)
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

func drawPanel(t *testing.T, panel *Panel, width int, height int) string {
	t.Helper()

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(screen.Fini)

	screen.SetSize(width, height)
	panel.list.SetRect(0, 0, width, height)
	panel.list.Draw(screen)
	screen.Show()

	cells, screenWidth, screenHeight := screen.GetContents()
	var rendered strings.Builder
	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			cell := cells[y*screenWidth+x]
			if len(cell.Runes) == 0 {
				rendered.WriteRune(' ')
				continue
			}
			rendered.WriteRune(cell.Runes[0])
		}
		rendered.WriteRune('\n')
	}

	return rendered.String()
}
