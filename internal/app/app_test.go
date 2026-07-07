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
