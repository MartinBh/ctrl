package battery

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/probes"
	batteryprobe "github.com/martinbhatta/ctrl/internal/probes/battery"
	"github.com/martinbhatta/ctrl/internal/theme"
)

type Panel struct {
	table *tview.Table
}

func NewPanel() *Panel {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	theme.Box(table.Box, "BATTERY")

	return &Panel{table: table}
}

func (p *Panel) Primitive() tview.Primitive {
	return p.table
}

func (p *Panel) SetStatus(status batteryprobe.Status) {
	p.table.Clear()

	color := statusColor(status.Level)
	p.table.SetCell(0, 0, tview.NewTableCell("Status").SetTextColor(theme.ColorPrimary))
	p.table.SetCell(0, 1, tview.NewTableCell(statusText(status)).SetTextColor(color))

	if status.Detail != "" {
		p.table.SetCell(1, 0, tview.NewTableCell("Detail").SetTextColor(theme.ColorPrimary))
		p.table.SetCell(1, 1, tview.NewTableCell(status.Detail).SetTextColor(theme.ColorMuted))
	}
}

func statusText(status batteryprobe.Status) string {
	if !status.Present {
		return status.State
	}

	return fmt.Sprintf("%.0f%% %s", status.Percent, status.State)
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
