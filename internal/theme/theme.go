package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const AppTitle = "ctrl"

var (
	ColorPrimary = tcell.ColorDarkCyan
	ColorAccent  = tcell.ColorGreen
	ColorMuted   = tcell.ColorGray
	ColorWarning = tcell.ColorYellow
	ColorError   = tcell.ColorRed
)

func Box(box *tview.Box, title string) {
	box.SetBorder(true)
	box.SetTitle(" " + title + " ")
	box.SetBorderColor(ColorPrimary)
	box.SetTitleColor(ColorAccent)
}
