package usage

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/format"
	"github.com/martinbhatta/ctrl/internal/probes"
	usageprobe "github.com/martinbhatta/ctrl/internal/probes/usage"
	"github.com/martinbhatta/ctrl/internal/theme"
)

type Panel struct {
	table *tview.Table
}

func NewPanel() *Panel {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	theme.Box(table.Box, "USAGE")

	return &Panel{table: table}
}

func (p *Panel) Primitive() tview.Primitive {
	return p.table
}

func (p *Panel) SetRows(rows []usageprobe.ResourceUsage) {
	p.table.Clear()

	p.table.SetCell(0, 0, headerCell("Resource"))
	p.table.SetCell(0, 1, headerCell("Used / Total"))
	p.table.SetCell(0, 2, headerCell("Free"))
	p.table.SetCell(0, 3, headerCell("Usage"))

	if len(rows) == 0 {
		p.table.SetCell(1, 0, tview.NewTableCell("No usage data").SetTextColor(theme.ColorMuted))
		return
	}

	for index, row := range rows {
		tableRow := index + 1
		color := statusColor(row.Level)

		p.table.SetCell(tableRow, 0, tview.NewTableCell(row.Name).SetTextColor(theme.ColorAccent))
		if row.Err != nil {
			p.table.SetCell(tableRow, 1, tview.NewTableCell("error").SetTextColor(color))
			p.table.SetCell(tableRow, 2, tview.NewTableCell(row.Detail).SetTextColor(theme.ColorMuted))
			continue
		}

		p.table.SetCell(tableRow, 1, tview.NewTableCell(fmt.Sprintf("%s / %s", format.Bytes(row.UsedBytes), format.Bytes(row.TotalBytes))).SetTextColor(color))
		p.table.SetCell(tableRow, 2, tview.NewTableCell(format.Bytes(row.FreeBytes)).SetTextColor(theme.ColorMuted))
		p.table.SetCell(tableRow, 3, tview.NewTableCell(fmt.Sprintf("%.0f%% %s", row.UsedPct, progressBar(row.UsedPct, 12))).SetTextColor(color))
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

func progressBar(usedPct float64, width int) string {
	if width <= 0 {
		return ""
	}

	usedPct = math.Max(0, math.Min(100, usedPct))
	filled := int(math.Round((usedPct / 100) * float64(width)))

	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}
