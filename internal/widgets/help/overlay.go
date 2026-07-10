package help

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/theme"
)

const (
	maxOverlayWidth  = 72
	minOverlayWidth  = 12
	minOverlayHeight = 6
)

const dismissalPrompt = "[::b]Press Enter, Escape, or q to start.[::-]"

const tutorialText = `[::b]Welcome to ctrl[::-]

[green]Todos[::-]
  up/down move | a add | e edit
  space complete/reopen | d delete

[green]Dashboard[::-]
  r refresh | ? help | q/Ctrl+C quit
  w weather | t todos

[gray]Todos save locally as you work.[::-]

` + dismissalPrompt

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
	if width < minOverlayWidth {
		width = screenWidth
	}
	if width <= 0 {
		return
	}

	innerWidth := width - 4
	if innerWidth < 1 {
		innerWidth = width
	}

	lines := tview.WordWrap(o.text, innerWidth)
	height := fit(len(lines)+2, screenHeight-2)
	if height < minOverlayHeight {
		height = screenHeight
	}
	if height <= 0 {
		return
	}

	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2

	o.Box.SetRect(x, y, width, height)
	o.Box.DrawForSubclass(screen, o)

	innerHeight := height - 2
	for row, line := range visibleLines(lines, innerHeight) {
		tview.Print(screen, strings.TrimRight(line, " "), x+2, y+1+row, innerWidth, tview.AlignLeft, tcell.ColorWhite)
	}
}

func fit(maximum, available int) int {
	if available < maximum {
		return available
	}

	return maximum
}

func visibleLines(lines []string, height int) []string {
	if height <= 0 {
		return nil
	}
	if len(lines) <= height {
		return lines
	}

	visible := append([]string(nil), lines[:height]...)
	visible[len(visible)-1] = dismissalPrompt
	return visible
}
