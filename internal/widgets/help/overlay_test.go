package help

import (
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
