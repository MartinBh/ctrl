package help

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestOverlayDrawsAtCommonTerminalSizes(t *testing.T) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{name: "standard", width: 100, height: 32},
		{name: "compact", width: 48, height: 16},
	}

	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			screen := tcell.NewSimulationScreen("")
			if err := screen.Init(); err != nil {
				t.Fatalf("Init() error = %v", err)
			}
			defer screen.Fini()

			screen.SetSize(size.width, size.height)

			overlay := NewOverlay()
			overlay.Draw(screen)
		})
	}
}

func TestOverlayShowsEssentialTextAtCompactSize(t *testing.T) {
	screen := drawOverlay(t, 48, 16)
	rendered := screenText(screen)

	for _, want := range []string{
		"Welcome to ctrl",
		"up/down move",
		"a add",
		"e edit",
		"space complete/reopen",
		"d delete",
		"r refresh",
		"? help",
		"q/Ctrl+C quit",
		"Todos save locally",
		"Press Enter",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("compact overlay missing %q in:\n%s", want, rendered)
		}
	}
}

func TestOverlayKeepsDismissalPromptVisibleWhenClipped(t *testing.T) {
	screen := drawOverlay(t, 36, 8)
	rendered := screenText(screen)

	if !strings.Contains(rendered, "Press Enter") {
		t.Fatalf("clipped overlay missing dismissal prompt in:\n%s", rendered)
	}
}

func drawOverlay(t *testing.T, width int, height int) tcell.SimulationScreen {
	t.Helper()

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	t.Cleanup(screen.Fini)

	screen.SetSize(width, height)
	NewOverlay().Draw(screen)
	screen.Show()

	return screen
}

func screenText(screen tcell.SimulationScreen) string {
	cells, width, height := screen.GetContents()

	var rendered strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := cells[y*width+x]
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
