package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/probes"
	usageprobe "github.com/martinbhatta/ctrl/internal/probes/usage"
	"github.com/martinbhatta/ctrl/internal/store"
	"github.com/martinbhatta/ctrl/internal/theme"
	envwidget "github.com/martinbhatta/ctrl/internal/widgets/env"
	todowidget "github.com/martinbhatta/ctrl/internal/widgets/todo"
	usagewidget "github.com/martinbhatta/ctrl/internal/widgets/usage"
)

type Options struct {
	Version      string
	TodoPath     string
	RefreshEvery time.Duration
}

type Dashboard struct {
	options Options

	app       *tview.Application
	todos     *todowidget.Panel
	env       *envwidget.Panel
	usage     *usagewidget.Panel
	footer    *tview.TextView
	todoStore *store.TodoStore
	probes    []probes.Probe
}

func New(options Options) *Dashboard {
	if options.RefreshEvery <= 0 {
		options.RefreshEvery = 5 * time.Minute
	}

	return &Dashboard{
		options:   options,
		app:       tview.NewApplication(),
		todos:     todowidget.NewPanel(),
		env:       envwidget.NewPanel(),
		usage:     usagewidget.NewPanel(),
		footer:    tview.NewTextView().SetDynamicColors(true),
		todoStore: store.NewTodoStore(options.TodoPath),
		probes:    probes.Default(),
	}
}

func (d *Dashboard) Run(ctx context.Context) error {
	d.configure()
	d.refreshSync(ctx)

	go d.refreshLoop(ctx)
	go func() {
		<-ctx.Done()
		d.app.Stop()
	}()

	return d.app.Run()
}

func (d *Dashboard) configure() {
	root := tview.NewFlex().SetDirection(tview.FlexRow)

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[::b]%s[::-]  [gray]personal terminal command center  version %s", theme.AppTitle, d.options.Version))

	body := tview.NewFlex().SetDirection(tview.FlexColumn)
	sidebar := tview.NewFlex().SetDirection(tview.FlexRow)
	sidebar.AddItem(d.env.Primitive(), 0, 2, false)
	sidebar.AddItem(d.usage.Primitive(), 0, 1, false)

	body.AddItem(d.todos.Primitive(), 0, 1, true)
	body.AddItem(sidebar, 0, 1, false)

	d.footer.SetTextColor(theme.ColorMuted)
	d.footer.SetTextAlign(tview.AlignCenter)
	d.setFooter("q quit | r refresh | data " + d.options.TodoPath)

	root.AddItem(header, 1, 0, false)
	root.AddItem(body, 0, 1, true)
	root.AddItem(d.footer, 1, 0, false)

	d.app.SetRoot(root, true)
	d.app.EnableMouse(true)
	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			d.app.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				d.app.Stop()
				return nil
			case 'r':
				d.setFooter("refreshing dashboard...")
				go d.refreshAsync(context.Background())
				return nil
			}
		}

		return event
	})
}

func (d *Dashboard) refreshSync(ctx context.Context) {
	todos, err := d.todoStore.Load()
	if err != nil {
		d.todos.SetError(err)
	} else {
		d.todos.SetTodos(todos)
	}

	statuses := probes.CheckAll(ctx, d.probes, 5*time.Second)
	d.env.SetStatuses(statuses)
	d.usage.SetRows(d.checkUsage(ctx))
}

func (d *Dashboard) refreshAsync(ctx context.Context) {
	todos, todoErr := d.todoStore.Load()
	statuses := probes.CheckAll(ctx, d.probes, 5*time.Second)
	usageRows := d.checkUsage(ctx)

	d.app.QueueUpdateDraw(func() {
		if todoErr != nil {
			d.todos.SetError(todoErr)
		} else {
			d.todos.SetTodos(todos)
		}
		d.env.SetStatuses(statuses)
		d.usage.SetRows(usageRows)
		d.setFooter("q quit | r refresh | data " + d.options.TodoPath)
	})
}

func (d *Dashboard) checkUsage(ctx context.Context) []usageprobe.ResourceUsage {
	usageCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return usageprobe.Check(usageCtx)
}

func (d *Dashboard) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(d.options.RefreshEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.refreshAsync(ctx)
		}
	}
}

func (d *Dashboard) setFooter(text string) {
	d.footer.SetText("[gray]" + text)
}
