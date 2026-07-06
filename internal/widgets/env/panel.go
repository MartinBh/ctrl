package env

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/probes"
	"github.com/martinbhatta/ctrl/internal/theme"
)

type Panel struct {
	table *tview.Table
}

func NewPanel() *Panel {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	theme.Box(table.Box, "ENVIRONMENT")

	return &Panel{table: table}
}

func (p *Panel) Primitive() tview.Primitive {
	return p.table
}

func (p *Panel) SetStatuses(statuses []probes.Status) {
	p.table.Clear()

	p.table.SetCell(0, 0, headerCell("Tool"))
	p.table.SetCell(0, 1, headerCell("Status"))
	p.table.SetCell(0, 2, headerCell("Detail"))

	for index, status := range statuses {
		row := index + 1
		color := statusColor(status.Level)

		p.table.SetCell(row, 0, tview.NewTableCell(status.Name).SetTextColor(theme.ColorAccent))
		p.table.SetCell(row, 1, tview.NewTableCell(status.Value).SetTextColor(color))
		p.table.SetCell(row, 2, tview.NewTableCell(status.Detail).SetTextColor(theme.ColorMuted))
	}
}

func headerCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(theme.ColorPrimary).
		SetSelectable(false).
		SetExpansion(1)
}

func statusColor(level probes.Level) tcell.Color {
	switch level {
	case probes.LevelOK:
		return theme.ColorAccent
	case probes.LevelWarning:
		return theme.ColorWarning
	case probes.LevelError:
		return theme.ColorError
	default:
		return theme.ColorMuted
	}
}
