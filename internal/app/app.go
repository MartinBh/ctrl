package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/probes"
	batteryprobe "github.com/martinbhatta/ctrl/internal/probes/battery"
	usageprobe "github.com/martinbhatta/ctrl/internal/probes/usage"
	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
	"github.com/martinbhatta/ctrl/internal/store"
	"github.com/martinbhatta/ctrl/internal/theme"
	batterywidget "github.com/martinbhatta/ctrl/internal/widgets/battery"
	envwidget "github.com/martinbhatta/ctrl/internal/widgets/env"
	helpwidget "github.com/martinbhatta/ctrl/internal/widgets/help"
	todowidget "github.com/martinbhatta/ctrl/internal/widgets/todo"
	usagewidget "github.com/martinbhatta/ctrl/internal/widgets/usage"
	weatherwidget "github.com/martinbhatta/ctrl/internal/widgets/weather"
)

const (
	dashboardPageName  = "dashboard"
	helpPageName       = "help"
	todoDeletePageName = "todo-delete"
	todoInputPageName  = "todo-input"
)

type Options struct {
	Version      string
	ConfigPath   string
	TodoPath     string
	RefreshEvery time.Duration
}

type Dashboard struct {
	options Options

	app                *tview.Application
	pages              *tview.Pages
	todos              *todowidget.Panel
	env                *envwidget.Panel
	usage              *usagewidget.Panel
	battery            *batterywidget.Panel
	weather            *weatherwidget.Panel
	footer             *tview.TextView
	config             store.Config
	configStore        *store.ConfigStore
	todoStore          *store.TodoStore
	probes             []probes.Probe
	weatherClient      weatherChecker
	helpVisible        bool
	todoOverlayVisible bool
}

type weatherChecker interface {
	Forecasts(context.Context) []weatherprobe.Forecast
}

func New(options Options) *Dashboard {
	if options.RefreshEvery <= 0 {
		options.RefreshEvery = 5 * time.Minute
	}

	return &Dashboard{
		options:       options,
		app:           tview.NewApplication(),
		todos:         todowidget.NewPanel(),
		env:           envwidget.NewPanel(),
		usage:         usagewidget.NewPanel(),
		battery:       batterywidget.NewPanel(),
		weather:       weatherwidget.NewPanel(),
		footer:        tview.NewTextView().SetDynamicColors(true),
		configStore:   store.NewConfigStore(options.ConfigPath),
		todoStore:     store.NewTodoStore(options.TodoPath),
		probes:        probes.Default(),
		weatherClient: weatherprobe.NewClient(),
	}
}

func (d *Dashboard) Run(ctx context.Context) error {
	config, configErr := d.configStore.Load()
	if configErr == nil {
		d.config = config
	}

	d.configure()

	if configErr != nil {
		d.setFooter("could not load config: " + configErr.Error())
	} else if !d.config.TutorialSeen {
		d.showHelp()
	}

	go d.refreshAsync(ctx, false)
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
	sidebar.AddItem(d.battery.Primitive(), 4, 0, false)
	sidebar.AddItem(d.weather.Primitive(), 0, 8, false)

	body.AddItem(d.todos.Primitive(), 0, 1, true)
	body.AddItem(sidebar, 0, 1, false)

	d.footer.SetTextColor(theme.ColorMuted)
	d.footer.SetTextAlign(tview.AlignCenter)
	d.setFooter(d.defaultFooter())

	root.AddItem(header, 1, 0, false)
	root.AddItem(body, 0, 1, true)
	root.AddItem(d.footer, 1, 0, false)

	d.pages = tview.NewPages()
	d.pages.AddPage(dashboardPageName, root, true, true)

	d.app.SetRoot(d.pages, true)
	d.app.EnableMouse(true)
	d.app.SetInputCapture(d.handleKey)
	d.showLoadingState()
}

func (d *Dashboard) handleKey(event *tcell.EventKey) *tcell.EventKey {
	if d.todoOverlayVisible {
		switch event.Key() {
		case tcell.KeyEscape:
			d.closeTodoOverlay()
			return nil
		case tcell.KeyCtrlC:
			d.app.Stop()
			return nil
		}

		return event
	}

	if d.helpVisible {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyEscape:
			d.dismissHelp()
			return nil
		case tcell.KeyCtrlC:
			d.app.Stop()
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				d.dismissHelp()
			}
			return nil
		}

		return nil
	}

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
			go d.refreshAsync(context.Background(), true)
			return nil
		case 'a':
			d.showAddTodo()
			return nil
		case 'e':
			d.showEditTodo()
			return nil
		case 'd':
			d.showDeleteTodo()
			return nil
		case ' ':
			d.toggleSelectedTodo()
			return nil
		case '?':
			d.showHelp()
			return nil
		}
	}

	return event
}

func (d *Dashboard) showAddTodo() {
	d.showTodoInput("ADD TODO", "Title", "", "todo added", func(title string) error {
		todos, todo, err := d.todoStore.Add(title)
		if err != nil {
			return err
		}

		d.todos.SetTodos(todos)
		d.todos.SelectTodo(todo.ID)
		return nil
	})
}

func (d *Dashboard) showEditTodo() {
	todo, ok := d.todos.SelectedTodo()
	if !ok {
		d.setFooter("no todo selected")
		return
	}

	d.showTodoInput("EDIT TODO", "Title", todo.Title, "todo updated", func(title string) error {
		todos, err := d.todoStore.UpdateTitle(todo.ID, title)
		if err != nil {
			return err
		}

		d.todos.SetTodos(todos)
		d.todos.SelectTodo(todo.ID)
		return nil
	})
}

func (d *Dashboard) toggleSelectedTodo() {
	todo, ok := d.todos.SelectedTodo()
	if !ok {
		d.setFooter("no todo selected")
		return
	}

	todos, err := d.todoStore.Toggle(todo.ID)
	if err != nil {
		d.setFooter("could not toggle todo: " + err.Error())
		return
	}

	d.todos.SetTodos(todos)
	d.todos.SelectTodo(todo.ID)
	if todo.Done {
		d.setFooter("todo marked incomplete")
	} else {
		d.setFooter("todo completed")
	}
}

func (d *Dashboard) showDeleteTodo() {
	todo, ok := d.todos.SelectedTodo()
	if !ok {
		d.setFooter("no todo selected")
		return
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete todo?\n\n%s", todo.Title)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(_ int, label string) {
			if label != "Delete" {
				d.closeTodoOverlay()
				return
			}

			todos, err := d.todoStore.Delete(todo.ID)
			if err != nil {
				d.setFooter("could not delete todo: " + err.Error())
				return
			}

			d.todos.SetTodos(todos)
			d.closeTodoOverlay()
			d.setFooter("todo deleted")
		})
	theme.Box(modal.Box, "DELETE TODO")

	d.showTodoOverlay(todoDeletePageName, centeredPrimitive(modal, 52, 10), modal)
}

func (d *Dashboard) showTodoInput(title string, label string, initial string, success string, save func(string) error) {
	input := tview.NewInputField().
		SetLabel(label + ": ").
		SetText(initial).
		SetFieldWidth(40).
		SetAcceptanceFunc(func(_ string, lastChar rune) bool {
			return lastChar != '\n' && lastChar != '\r'
		})

	form := tview.NewForm().
		AddFormItem(input).
		AddButton("Save", func() {
			d.saveTodoInput(input.GetText(), success, save)
		}).
		AddButton("Cancel", d.closeTodoOverlay).
		SetButtonsAlign(tview.AlignRight).
		SetLabelColor(theme.ColorPrimary).
		SetFieldTextColor(theme.ColorAccent).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetButtonTextColor(theme.ColorPrimary).
		SetButtonBackgroundColor(tcell.ColorBlack)
	form.SetCancelFunc(d.closeTodoOverlay)
	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			d.saveTodoInput(input.GetText(), success, save)
		case tcell.KeyEscape:
			d.closeTodoOverlay()
		}
	})
	theme.Box(form.Box, title)

	d.showTodoOverlay(todoInputPageName, centeredPrimitive(form, 58, 7), form)
}

func (d *Dashboard) saveTodoInput(title string, success string, save func(string) error) {
	if err := save(title); err != nil {
		d.setFooter("could not save todo: " + err.Error())
		return
	}

	d.closeTodoOverlay()
	d.setFooter(success)
}

func (d *Dashboard) showTodoOverlay(pageName string, primitive tview.Primitive, focus tview.Primitive) {
	if d.pages == nil {
		return
	}

	d.closeTodoOverlay()
	d.todoOverlayVisible = true
	d.pages.AddPage(pageName, primitive, true, true)
	d.pages.SendToFront(pageName)
	d.app.SetFocus(focus)
}

func (d *Dashboard) closeTodoOverlay() {
	if d.pages == nil {
		return
	}

	d.todoOverlayVisible = false
	d.pages.RemovePage(todoInputPageName)
	d.pages.RemovePage(todoDeletePageName)
	d.app.SetFocus(d.todos.Primitive())
}

func (d *Dashboard) showHelp() {
	if d.pages == nil {
		return
	}

	d.helpVisible = true
	d.pages.AddPage(helpPageName, helpwidget.NewOverlay(), true, true)
	d.pages.SendToFront(helpPageName)
}

func (d *Dashboard) dismissHelp() {
	if !d.helpVisible {
		return
	}

	d.helpVisible = false
	d.pages.RemovePage(helpPageName)

	if !d.config.TutorialSeen {
		d.config.TutorialSeen = true
		if err := d.configStore.Save(d.config); err != nil {
			d.setFooter("could not save tutorial state: " + err.Error())
			return
		}
	}

	d.setFooter(d.defaultFooter())
}

func (d *Dashboard) refreshAsync(ctx context.Context, resetFooter bool) {
	statuses := probes.CheckAll(ctx, d.probes, 5*time.Second)
	usageRows := d.checkUsage(ctx)
	batteryStatus := d.checkBattery(ctx)
	weatherForecasts := d.checkWeather(ctx)

	d.app.QueueUpdateDraw(func() {
		d.refreshTodos()
		d.env.SetStatuses(statuses)
		d.usage.SetRows(usageRows)
		d.battery.SetStatus(batteryStatus)
		d.weather.SetForecasts(weatherForecasts)
		if resetFooter {
			d.setFooter(d.defaultFooter())
		}
	})
}

func (d *Dashboard) refreshTodos() {
	todos, err := d.todoStore.Load()
	if err != nil {
		d.todos.SetError(err)
		return
	}

	d.todos.SetTodos(todos)
}

func (d *Dashboard) checkUsage(ctx context.Context) []usageprobe.ResourceUsage {
	usageCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return usageprobe.Check(usageCtx)
}

func (d *Dashboard) checkBattery(ctx context.Context) batteryprobe.Status {
	batteryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return batteryprobe.Check(batteryCtx)
}

func (d *Dashboard) checkWeather(ctx context.Context) []weatherprobe.Forecast {
	weatherCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return d.weatherClient.Forecasts(weatherCtx)
}

func (d *Dashboard) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(d.options.RefreshEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.refreshAsync(ctx, true)
		}
	}
}

func (d *Dashboard) showLoadingState() {
	d.todos.SetLoading()
	d.env.SetStatuses(loadingProbeStatuses(d.probes))
	d.usage.SetLoading()
	d.battery.SetStatus(batteryprobe.Status{
		Present: false,
		State:   "checking",
		Detail:  "battery status pending",
		Level:   probes.LevelMuted,
	})
	d.weather.SetLoading()
}

func loadingProbeStatuses(checks []probes.Probe) []probes.Status {
	statuses := make([]probes.Status, len(checks))
	for index, check := range checks {
		statuses[index] = probes.Status{
			Name:   check.Name(),
			Value:  "checking",
			Detail: "probe pending",
			Level:  probes.LevelMuted,
		}
	}

	return statuses
}

func (d *Dashboard) setFooter(text string) {
	d.footer.SetText("[gray]" + text)
}

func (d *Dashboard) defaultFooter() string {
	return "a add | e edit | space complete | d delete | r refresh | ? help | q quit | Weather: Open-Meteo | data " + d.options.TodoPath
}

func centeredPrimitive(primitive tview.Primitive, width int, height int) tview.Primitive {
	row := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(primitive, width, 0, true).
		AddItem(nil, 0, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(row, height, 0, true).
		AddItem(nil, 0, 1, false)
}
