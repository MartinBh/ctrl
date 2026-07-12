package app

import (
	"testing"

	"github.com/rivo/tview"
)

func TestLayoutUsesTwoByTwoGridOnWideTerminals(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	layout := dashboard.layout
	layout.Update(120, 40)

	if layout.mode != layoutWide {
		t.Fatalf("layout mode = %v, want wide", layout.mode)
	}
	if got := layout.Main.GetItemCount(); got != 2 {
		t.Fatalf("wide main panel count = %d, want 2 rows", got)
	}

	top, ok := layout.Main.GetItem(0).(*tview.Flex)
	if !ok {
		t.Fatalf("wide top row = %T, want *tview.Flex", layout.Main.GetItem(0))
	}
	if got := top.GetItemCount(); got != 2 {
		t.Fatalf("wide top panel count = %d, want todo and environment", got)
	}
	if got := top.GetItem(0); got != dashboard.todos.Primitive() {
		t.Fatal("wide top-left panel is not the todo list")
	}
	if got := top.GetItem(1); got != dashboard.env.Primitive() {
		t.Fatal("wide top-right panel is not environment status")
	}
}

func TestLayoutStacksPanelsOnNarrowTerminals(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	layout := dashboard.layout
	layout.Update(80, 40)

	if layout.mode != layoutNarrow {
		t.Fatalf("layout mode = %v, want narrow", layout.mode)
	}
	if got := layout.Main.GetItemCount(); got != 6 {
		t.Fatalf("narrow panel count = %d, want 6 stacked panels", got)
	}
	if got := layout.Main.GetItem(0); got != dashboard.todos.Primitive() {
		t.Fatal("narrow first panel is not the todo list")
	}
	if got := layout.Main.GetItem(1); got != dashboard.env.Primitive() {
		t.Fatal("narrow second panel is not environment status")
	}
}

func TestLayoutShowsWarningBelowMinimumSize(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	layout := dashboard.layout
	layout.Update(minimumWidth-1, 40)
	if layout.mode != layoutTooSmall {
		t.Fatalf("layout mode = %v, want too small", layout.mode)
	}
	if got := layout.Main.GetItemCount(); got != 1 {
		t.Fatalf("small layout panel count = %d, want warning", got)
	}
	if got := layout.Main.GetItem(0); got != layout.warning {
		t.Fatal("small layout does not show the size warning")
	}

	layout.Update(80, minimumHeight-1)
	if layout.mode != layoutTooSmall {
		t.Fatalf("short layout mode = %v, want too small", layout.mode)
	}
}
