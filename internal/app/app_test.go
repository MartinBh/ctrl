package app

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/martinbhatta/ctrl/internal/probes"
	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
	"github.com/martinbhatta/ctrl/internal/store"
)

func TestQuestionMarkOpensHelp(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	event := tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(?) = %v, want nil", got)
	}

	if !dashboard.helpVisible {
		t.Fatal("helpVisible = false, want true")
	}
	if !dashboard.pages.HasPage(helpPageName) {
		t.Fatal("help page was not added")
	}
}

func TestWeatherShortcutFocusesForecastPane(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	if got := dashboard.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone)); got != nil {
		t.Fatalf("handleKey(w) = %v, want nil", got)
	}
	if got, want := dashboard.app.GetFocus(), dashboard.weather.Primitive(); got != want {
		t.Fatal("weather table was not focused")
	}
}

func TestTabReturnsFocusToTodoList(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	dashboard.app.SetFocus(dashboard.weather.Primitive())

	if got := dashboard.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)); got != nil {
		t.Fatalf("handleKey(Tab) = %v, want nil", got)
	}
	if got, want := dashboard.app.GetFocus(), dashboard.todos.Primitive(); got != want {
		t.Fatal("Tab did not restore focus to the todo list")
	}
}

func TestHelpConsumesDashboardShortcuts(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	dashboard.showHelp()

	event := tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(r) = %v, want nil", got)
	}

	if !dashboard.helpVisible {
		t.Fatal("helpVisible = false, want true")
	}
}

func TestDismissHelpPersistsTutorialSeen(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	dashboard := New(Options{
		Version:      "test",
		ConfigPath:   configPath,
		TodoPath:     filepath.Join(dir, "todos.json"),
		RefreshEvery: time.Minute,
	})
	dashboard.configure()
	dashboard.showHelp()

	dashboard.dismissHelp()

	if dashboard.helpVisible {
		t.Fatal("helpVisible = true, want false")
	}
	if dashboard.pages.HasPage(helpPageName) {
		t.Fatal("help page should be removed")
	}

	config, err := store.NewConfigStore(configPath).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !config.TutorialSeen {
		t.Fatal("TutorialSeen = false, want true")
	}
}

func TestConfigureSeedsLoadingState(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.probes = []probes.Probe{staticProbe{name: "Slow"}}

	dashboard.configure()

	todoList := dashboard.todos.Primitive().(*tview.List)
	main, _ := todoList.GetItemText(0)
	if main != "[gray]Loading todos..." {
		t.Fatalf("todo loading text = %q, want loading state", main)
	}

	envTable := dashboard.env.Primitive().(*tview.Table)
	if got := envTable.GetCell(1, 1).Text; got != "checking" {
		t.Fatalf("environment status = %q, want checking", got)
	}

	usageTable := dashboard.usage.Primitive().(*tview.Table)
	if got := usageTable.GetCell(1, 1).Text; got != "checking" {
		t.Fatalf("usage status = %q, want checking", got)
	}

	if got := dashboard.weather.Primitive(); got == nil {
		t.Fatal("weather loading state is empty")
	}

	batteryTable := dashboard.battery.Primitive().(*tview.Table)
	if got := batteryTable.GetCell(0, 1).Text; got != "checking" {
		t.Fatalf("battery status = %q, want checking", got)
	}
}

func TestRunStartsBeforeInitialRefreshCompletes(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.app.SetScreen(tcell.NewSimulationScreen(""))

	slowProbe := &blockingProbe{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	dashboard.probes = []probes.Probe{slowProbe}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- dashboard.Run(ctx)
	}()

	waitForSignal(t, slowProbe.started, "slow probe to start")

	rendered := make(chan struct{})
	dashboard.app.QueueUpdateDraw(func() {
		close(rendered)
	})
	waitForSignal(t, rendered, "application update to run")

	cancel()
	close(slowProbe.release)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		dashboard.app.Stop()
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestRefreshStartsWeatherBeforeSlowProbeCompletes(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.app.SetScreen(tcell.NewSimulationScreen(""))

	slowProbe := &blockingProbe{started: make(chan struct{}), release: make(chan struct{})}
	slowWeather := &blockingWeatherChecker{started: make(chan struct{}), release: make(chan struct{})}
	dashboard.probes = []probes.Probe{slowProbe}
	dashboard.weatherClient = slowWeather

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- dashboard.Run(ctx) }()

	waitForSignal(t, slowProbe.started, "slow probe to start")
	waitForSignal(t, slowWeather.started, "weather refresh to start")

	close(slowProbe.release)
	close(slowWeather.release)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		dashboard.app.Stop()
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestRefreshKeepsTodoChangesMadeDuringSlowProbe(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.app.SetScreen(tcell.NewSimulationScreen(""))

	todos := []store.Todo{{ID: "todo-1", Title: "Old title"}}
	if err := dashboard.todoStore.Save(todos); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	slowProbe := &blockingProbe{
		started: make(chan struct{}),
		release: make(chan struct{}),
		done:    make(chan struct{}),
	}
	dashboard.probes = []probes.Probe{slowProbe}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- dashboard.Run(ctx)
	}()

	waitForSignal(t, slowProbe.started, "slow probe to start")

	updated, err := dashboard.todoStore.UpdateTitle("todo-1", "Fresh title")
	if err != nil {
		t.Fatalf("UpdateTitle() error = %v", err)
	}
	queueAndWait(t, dashboard, func() {
		dashboard.todos.SetTodos(updated)
	})

	close(slowProbe.release)
	waitForSignal(t, slowProbe.done, "slow probe to finish")
	waitForTodoAndProbe(t, dashboard, "Fresh title", "ok")

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		dashboard.app.Stop()
		t.Fatal("Run() did not stop after context cancellation")
	}
}

func TestAddShortcutOpensAndEscapeClosesTodoInput(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()

	event := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(a) = %v, want nil", got)
	}

	if !dashboard.todoOverlayVisible {
		t.Fatal("todoOverlayVisible = false, want true")
	}
	if !dashboard.pages.HasPage(todoInputPageName) {
		t.Fatal("todo input page was not added")
	}

	event = tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(Escape) = %v, want nil", got)
	}

	if dashboard.todoOverlayVisible {
		t.Fatal("todoOverlayVisible = true, want false")
	}
	if dashboard.pages.HasPage(todoInputPageName) {
		t.Fatal("todo input page should be removed")
	}
}

func TestSpaceTogglesSelectedTodo(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	todos := []store.Todo{{ID: "todo-1", Title: "Write tests"}}
	if err := dashboard.todoStore.Save(todos); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	dashboard.todos.SetTodos(todos)

	event := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(space) = %v, want nil", got)
	}

	loaded, err := dashboard.todoStore.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded) != 1 || !loaded[0].Done {
		t.Fatalf("Load() = %#v, want completed todo", loaded)
	}
}

func TestTodoActionsWithoutSelectionShowFooterMessage(t *testing.T) {
	dashboard := testDashboard(t)
	dashboard.configure()
	dashboard.todos.SetTodos(nil)

	event := tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)
	if got := dashboard.handleKey(event); got != nil {
		t.Fatalf("handleKey(e) = %v, want nil", got)
	}

	if got := dashboard.footer.GetText(true); got != "no todo selected" {
		t.Fatalf("footer = %q, want %q", got, "no todo selected")
	}
}

func testDashboard(t *testing.T) *Dashboard {
	t.Helper()

	dir := t.TempDir()
	dashboard := New(Options{
		Version:      "test",
		ConfigPath:   filepath.Join(dir, "config.json"),
		TodoPath:     filepath.Join(dir, "todos.json"),
		RefreshEvery: time.Minute,
	})
	dashboard.weatherClient = staticWeatherChecker{}
	return dashboard
}

type staticWeatherChecker struct{}

func (staticWeatherChecker) Forecasts(context.Context) []weatherprobe.Forecast {
	return []weatherprobe.Forecast{{Location: weatherprobe.Location{Name: "Local weather"}}}
}

type staticProbe struct {
	name string
}

func (p staticProbe) Name() string {
	return p.name
}

func (p staticProbe) Check(context.Context) probes.Status {
	return probes.Status{Name: p.name, Value: "ok", Level: probes.LevelOK}
}

type blockingProbe struct {
	started chan struct{}
	release chan struct{}
	done    chan struct{}
}

type blockingWeatherChecker struct {
	started chan struct{}
	release chan struct{}
}

func (c *blockingWeatherChecker) Forecasts(ctx context.Context) []weatherprobe.Forecast {
	close(c.started)
	select {
	case <-c.release:
		return []weatherprobe.Forecast{{Location: weatherprobe.Location{Name: "Local weather"}}}
	case <-ctx.Done():
		return []weatherprobe.Forecast{{Location: weatherprobe.Location{Name: "Local weather"}, Err: ctx.Err()}}
	}
}

func (p *blockingProbe) Name() string {
	return "Slow"
}

func (p *blockingProbe) Check(ctx context.Context) probes.Status {
	close(p.started)
	if p.done != nil {
		defer close(p.done)
	}

	select {
	case <-p.release:
		return probes.Status{Name: p.Name(), Value: "ok", Level: probes.LevelOK}
	case <-ctx.Done():
		return probes.Status{Name: p.Name(), Value: "timeout", Detail: ctx.Err().Error(), Level: probes.LevelError}
	}
}

func queueAndWait(t *testing.T, dashboard *Dashboard, update func()) {
	t.Helper()

	done := make(chan struct{})
	dashboard.app.QueueUpdateDraw(func() {
		update()
		close(done)
	})
	waitForSignal(t, done, "queued update to run")
}

func waitForTodoAndProbe(t *testing.T, dashboard *Dashboard, todoTitle string, probeValue string) {
	t.Helper()

	deadline := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for todo %q and probe %q", todoTitle, probeValue)
		case <-ticker.C:
			state := readDashboardState(t, dashboard)
			if strings.Contains(state.todo, todoTitle) && state.probe == probeValue {
				return
			}
		}
	}
}

type dashboardState struct {
	todo  string
	probe string
}

func readDashboardState(t *testing.T, dashboard *Dashboard) dashboardState {
	t.Helper()

	ch := make(chan dashboardState, 1)
	dashboard.app.QueueUpdateDraw(func() {
		todoList := dashboard.todos.Primitive().(*tview.List)
		todo, _ := todoList.GetItemText(0)

		envTable := dashboard.env.Primitive().(*tview.Table)
		probe := envTable.GetCell(1, 1).Text

		ch <- dashboardState{todo: todo, probe: probe}
	})

	select {
	case state := <-ch:
		return state
	case <-time.After(2 * time.Second):
		t.Fatal("timed out reading dashboard state")
		return dashboardState{}
	}
}

func waitForSignal(t *testing.T, ch <-chan struct{}, name string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", name)
	}
}
