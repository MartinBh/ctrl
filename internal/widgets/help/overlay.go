package help

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/theme"
)

const (
	maxOverlayWidth  = 72
	maxOverlayHeight = 20
)

const tutorialText = `[::b]Welcome to ctrl[::-]

ctrl is your personal terminal command center.

[green]Todo panel[::-]
  up/down  move through visible todos
  enter    select the highlighted row

[green]Dashboard controls[::-]
  r        refresh environment and system status
  ?        show this help screen again
  q        quit
  Ctrl+C   quit

[gray]Todos load from your local JSON file. Add, complete, and delete controls are coming in the interactive todo pass.[::-]

[::b]Press Enter, Escape, or q to start.[::-]`

type Overlay struct {
	*tview.Box
	text string
}

func NewOverlay() *Overlay {
	box := tview.NewBox()
	theme.Box(box, "HELP")
	box.SetBackgroundColor(tcell.ColorBlack)

	return &Overlay{
		Box:  box,
		text: tutorialText,
	}
}

func Text() string {
	return tutorialText
}

func (o *Overlay) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	if screenWidth <= 0 || screenHeight <= 0 {
		return
	}

	width := fit(maxOverlayWidth, screenWidth-2)
	height := fit(maxOverlayHeight, screenHeight-2)
	if width < 12 {
		width = screenWidth
	}
	if height < 6 {
		height = screenHeight
	}

	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2

	o.Box.SetRect(x, y, width, height)
	o.Box.DrawForSubclass(screen, o)

	lines := tview.WordWrap(o.text, width-4)
	innerHeight := height - 2
	for row, line := range lines {
		if row >= innerHeight {
			break
		}
		tview.Print(screen, strings.TrimRight(line, " "), x+2, y+1+row, width-4, tview.AlignLeft, tcell.ColorWhite)
	}
}

func fit(maximum, available int) int {
	if available < maximum {
		return available
	}

	return maximum
}
