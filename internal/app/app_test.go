package app

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/martinbhatta/ctrl/internal/store"
)

func TestQuestionMarkOpensHelp(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	event := tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(?) = %v, want nil", got)
	}

	if !dashboard.helpVisible {
		t.Fatal("helpVisible = false, want true")
	}
	if !dashboard.pages.HasPage(helpPageName) {
		t.Fatal("help page was not added")
	}
}

func TestHelpConsumesDashboardShortcuts(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	dashboard.showHelp()

	event := tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(r) = %v, want nil", got)
	}

	if !dashboard.helpVisible {
		t.Fatal("helpVisible = false, want true")
	}
}

func TestDismissHelpPersistsTutorialSeen(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	dashboard := New(Options{
		Version:      "test",
		ConfigPath:   configPath,
		TodoPath:     filepath.Join(dir, "todos.json"),
		RefreshEvery: time.Minute,
	})
	dashboard.configure()
	dashboard.showHelp()

	dashboard.dismissHelp()

	if dashboard.helpVisible {
		t.Fatal("helpVisible = true, want false")
	}
	if dashboard.pages.HasPage(helpPageName) {
		t.Fatal("help page should be removed")
	}

	config, err := store.NewConfigStore(configPath).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !config.TutorialSeen {
		t.Fatal("TutorialSeen = false, want true")
	}
}

func TestAddShortcutOpensAndEscapeClosesTodoInput(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	event := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(a) = %v, want nil", got)
	}

	if !dashboard.todoOverlayVisible {
		t.Fatal("todoOverlayVisible = false, want true")
	}
	if !dashboard.pages.HasPage(todoInputPageName) {
		t.Fatal("todo input page was not added")
	}

	event = tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(Escape) = %v, want nil", got)
	}

	if dashboard.todoOverlayVisible {
		t.Fatal("todoOverlayVisible = true, want false")
	}
	if dashboard.pages.HasPage(todoInputPageName) {
		t.Fatal("todo input page should be removed")
	}
}

func TestSpaceTogglesSelectedTodo(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	todos := []store.Todo{{ID: "todo-1", Title: "Write tests"}}
	if err := dashboard.todoStore.Save(todos); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	dashboard.todos.SetTodos(todos)

	event := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(space) = %v, want nil", got)
	}

	loaded, err := dashboard.todoStore.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 1 || !loaded[0].Done {
		t.Fatalf("Load() = %#v, want completed todo", loaded)
	}
}

func TestTodoActionsWithoutSelectionShowFooterMessage(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	dashboard.todos.SetTodos(nil)

	event := tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(e) = %v, want nil", got)
	}

	if got := dashboard.footer.GetText(true); got != "no todo selected" {
		t.Fatalf("footer = %q, want %q", got, "no todo selected")
	}
}

func testDashboard(t *testing.T) *Dashboard {
	t.Helper()

	dir := t.TempDir()
	return New(Options{
		Version:      "test",
		ConfigPath:   filepath.Join(dir, "config.json"),
		TodoPath:     filepath.Join(dir, "todos.json"),
		RefreshEvery: time.Minute,
	})
}
