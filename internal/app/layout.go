package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/theme"
)

const (
	wideLayoutWidth = 100
	minimumWidth    = 52
	minimumHeight   = 16
)

type layoutMode int

const (
	layoutWide layoutMode = iota
	layoutNarrow
	layoutTooSmall
)

// Layout composes the dashboard panels and selects a usable arrangement for
// the available terminal size. Pages is intentionally the root so overlays
// can sit above the dashboard without replacing it.
type Layout struct {
	Pages  *tview.Pages
	Root   *tview.Flex
	Main   *tview.Flex
	Header *tview.TextView
	Footer *tview.TextView

	todos   tview.Primitive
	env     tview.Primitive
	usage   tview.Primitive
	battery tview.Primitive
	weather tview.Primitive
	status  tview.Primitive
	warning *tview.TextView
	mode    layoutMode
}

func newLayout(version string, todos, env, usage, battery, weather tview.Primitive, footer *tview.TextView) *Layout {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::b]" + theme.AppTitle + "[::-]  [gray]personal terminal command center  version " + version)

	status := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]HELP / STATUS[-]\n\n[gray]q[-] quit  [gray]?[-] help  [gray]r[-] refresh\n[gray]a[-] add todo  [gray]space[-] toggle  [gray]d[-] delete\n\n[gray]Tab returns to the todo list.\nWeather data: Open-Meteo[-]")
	theme.Box(status.Box, "STATUS")

	warning := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]Terminal is too small for the dashboard.[-]\nResize to at least 52 columns by 16 rows.")
	theme.Box(warning.Box, "CTRL")

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(header, 1, 0, false)
	root.AddItem(main, 0, 1, true)
	root.AddItem(footer, 1, 0, false)

	layout := &Layout{
		Root:    root,
		Main:    main,
		Header:  header,
		Footer:  footer,
		todos:   todos,
		env:     env,
		usage:   usage,
		battery: battery,
		weather: weather,
		status:  status,
		warning: warning,
	}
	layout.Pages = tview.NewPages().AddPage(dashboardPageName, root, true, true)
	layout.Update(wideLayoutWidth, minimumHeight)
	return layout
}

// Update changes the layout only when a terminal crosses a breakpoint.
func (l *Layout) Update(width, height int) {
	next := layoutForSize(width, height)
	if next == l.mode && l.Main.GetItemCount() > 0 {
		return
	}

	l.mode = next
	l.Main.Clear()

	switch next {
	case layoutWide:
		l.addWidePanels()
	case layoutNarrow:
		l.addNarrowPanels()
	case layoutTooSmall:
		l.Main.SetDirection(tview.FlexRow)
		l.Main.AddItem(l.warning, 0, 1, false)
	}
}

func layoutForSize(width, height int) layoutMode {
	if width < minimumWidth || height < minimumHeight {
		return layoutTooSmall
	}
	if width < wideLayoutWidth {
		return layoutNarrow
	}
	return layoutWide
}

func (l *Layout) addWidePanels() {
	l.Main.SetDirection(tview.FlexRow)

	top := tview.NewFlex().
		AddItem(l.todos, 0, 1, true).
		AddItem(l.env, 0, 1, false)

	usageAndBattery := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(l.usage, 0, 2, false).
		AddItem(l.battery, 0, 1, false)
	statusAndWeather := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(l.status, 0, 1, false).
		AddItem(l.weather, 0, 2, false)
	bottom := tview.NewFlex().
		AddItem(usageAndBattery, 0, 1, false).
		AddItem(statusAndWeather, 0, 1, false)

	l.Main.AddItem(top, 0, 1, true)
	l.Main.AddItem(bottom, 0, 1, false)
}

func (l *Layout) addNarrowPanels() {
	l.Main.SetDirection(tview.FlexRow)
	l.Main.AddItem(l.todos, 0, 2, true)
	l.Main.AddItem(l.env, 0, 1, false)
	l.Main.AddItem(l.usage, 0, 1, false)
	l.Main.AddItem(l.battery, 0, 1, false)
	l.Main.AddItem(l.status, 0, 1, false)
	l.Main.AddItem(l.weather, 0, 2, false)
}

func (l *Layout) BeforeDraw(screen tcell.Screen) bool {
	width, height := screen.Size()
	l.Update(width, height)
	return false
}
