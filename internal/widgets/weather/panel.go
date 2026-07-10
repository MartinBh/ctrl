package weather

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
	"github.com/martinbhatta/ctrl/internal/theme"
)

const compactLayoutWidth = 90

type Panel struct {
	root           *tview.Flex
	status         *tview.TextView
	cards          *tview.Flex
	detailTitle    *tview.TextView
	periods        *tview.Flex
	daily          *tview.TextView
	forecasts      []weatherprobe.Forecast
	lastSuccessful map[string]weatherprobe.Forecast
	stale          map[string]bool
	active         int
	narrow         bool
}

func NewPanel() *Panel {
	p := &Panel{
		status:         textView(),
		cards:          tview.NewFlex(),
		detailTitle:    textView(),
		periods:        tview.NewFlex(),
		daily:          textView(),
		lastSuccessful: make(map[string]weatherprobe.Forecast),
		stale:          make(map[string]bool),
	}
	p.root = tview.NewFlex().SetDirection(tview.FlexRow)
	p.root.AddItem(p.status, 1, 0, false)
	p.root.AddItem(p.cards, 6, 0, false)
	p.root.AddItem(p.detailTitle, 1, 0, false)
	p.root.AddItem(p.periods, 6, 0, false)
	p.root.AddItem(p.daily, 0, 1, false)
	p.root.SetInputCapture(p.handleKey)
	p.root.SetDrawFunc(func(_ tcell.Screen, x, y, width, height int) (int, int, int, int) {
		p.adaptLayout(width)
		return x + 1, y + 1, max(0, width-2), max(0, height-2)
	})
	theme.Box(p.root.Box, "WEATHER")
	p.SetLoading()

	return p
}

func (p *Panel) Primitive() tview.Primitive {
	return p.root
}

func (p *Panel) SetLoading() {
	p.forecasts = nil
	p.active = 0
	p.status.SetText("[gray]Open-Meteo · loading weather for Gangnam-gu and Sangbong-dong[-]")
	p.cards.Clear()
	for _, location := range weatherprobe.Locations {
		card := cardView(location.Name)
		card.SetText("[gray]Loading current conditions...[-]")
		p.cards.AddItem(card, 0, 1, false)
	}
	p.detailTitle.SetText("[gray]Forecast detail will appear after loading.[-]")
	p.periods.Clear()
	p.daily.SetText("")
}

func (p *Panel) SetForecasts(forecasts []weatherprobe.Forecast) {
	selectedLocation := p.selectedLocation()
	p.forecasts = make([]weatherprobe.Forecast, 0, len(forecasts))
	p.stale = make(map[string]bool)

	for _, forecast := range forecasts {
		if forecast.Err != nil {
			if previous, ok := p.lastSuccessful[forecast.Location.Name]; ok {
				p.forecasts = append(p.forecasts, previous)
				p.stale[forecast.Location.Name] = true
				continue
			}
			p.forecasts = append(p.forecasts, forecast)
			continue
		}

		p.lastSuccessful[forecast.Location.Name] = forecast
		p.forecasts = append(p.forecasts, forecast)
	}

	p.active = indexForLocation(p.forecasts, selectedLocation)
	p.render()
}

func (p *Panel) SetActiveLocation(index int) bool {
	if index < 0 || index >= len(p.forecasts) {
		return false
	}
	p.active = index
	p.render()
	return true
}

func (p *Panel) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyLeft, tcell.KeyUp:
		p.moveActive(-1)
		return nil
	case tcell.KeyRight, tcell.KeyDown, tcell.KeyTAB:
		p.moveActive(1)
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case '1':
			p.SetActiveLocation(0)
			return nil
		case '2':
			p.SetActiveLocation(1)
			return nil
		}
	}

	return event
}

func (p *Panel) moveActive(delta int) {
	if len(p.forecasts) == 0 {
		return
	}
	p.active = (p.active + delta + len(p.forecasts)) % len(p.forecasts)
	p.render()
}

func (p *Panel) adaptLayout(width int) {
	narrow := width < compactLayoutWidth
	if narrow == p.narrow {
		return
	}
	p.narrow = narrow

	if narrow {
		p.cards.SetDirection(tview.FlexRow)
		p.root.ResizeItem(p.cards, 12, 0)
		return
	}

	p.cards.SetDirection(tview.FlexColumn)
	p.root.ResizeItem(p.cards, 6, 0)
}

func (p *Panel) render() {
	p.cards.Clear()
	for _, forecast := range p.forecasts {
		card := cardView(forecast.Location.Name)
		card.SetText(p.cardText(forecast))
		p.cards.AddItem(card, 0, 1, false)
	}

	if len(p.forecasts) == 0 {
		p.detailTitle.SetText("[gray]Weather is unavailable.[-]")
		p.periods.Clear()
		p.daily.SetText("")
		return
	}
	if p.active >= len(p.forecasts) {
		p.active = 0
	}

	forecast := p.forecasts[p.active]
	if forecast.Err != nil {
		p.detailTitle.SetText(fmt.Sprintf("[red]%s forecast unavailable: %s[-]", forecast.Location.Name, tview.Escape(forecast.Err.Error())))
		p.periods.Clear()
		p.daily.SetText("")
		return
	}

	p.detailTitle.SetText(fmt.Sprintf("[turquoise]%s forecast[-]  [gray]1/2 or left/right changes location · t returns to todos[-]", forecast.Location.Name))
	p.renderPeriods(forecast.Hourly)
	p.daily.SetText(p.dailyText(forecast.Daily))
}

func (p *Panel) cardText(forecast weatherprobe.Forecast) string {
	if forecast.Err != nil {
		return fmt.Sprintf("[red]Unavailable[-]\n[gray]%s[-]", tview.Escape(forecast.Err.Error()))
	}

	visual := weatherprobe.ConditionVisual(forecast.Current.WeatherCode)
	stale := ""
	if p.stale[forecast.Location.Name] {
		stale = " [red]STALE[-]"
	}
	return fmt.Sprintf("[%s]%s[-]  [%s]%-14s[-]%s\n[%s]%s[-]  [green]%.0f°C[-] [gray]feels %.0f°C[-]\n[%s]%s[-]  [blue]rain %.1f mm[-] [gray]· %.0f%% humidity[-]\n[gray]wind %.0f km/h %s[-]",
		visual.Color, visual.Glyph[0], visual.Color, strings.ToUpper(visual.Label), stale,
		visual.Color, visual.Glyph[1], forecast.Current.Temperature, forecast.Current.ApparentTemperature,
		visual.Color, visual.Glyph[2], forecast.Current.Precipitation, forecast.Current.Humidity,
		forecast.Current.WindSpeed, weatherprobe.WindDirection(forecast.Current.WindDirection))
}

func (p *Panel) renderPeriods(hourly []weatherprobe.Hourly) {
	p.periods.Clear()
	periods := weatherprobe.SummarizePeriods(hourly)
	if len(periods) == 0 {
		p.periods.AddItem(textView().SetText("[gray]No upcoming hourly forecast is available.[-]"), 0, 1, false)
		return
	}

	for _, period := range periods {
		visual := weatherprobe.ConditionVisual(period.Condition)
		card := textView()
		theme.Box(card.Box, strings.ToUpper(period.Label))
		card.SetText(fmt.Sprintf("[%s]%s[-] %s\n[green]%.0f–%.0f°C[-]\n[blue]rain %.0f%%[-]\n[gray]wind %.0f km/h[-]",
			visual.Color, visual.Glyph[1], visual.Label, period.Low, period.High, period.PrecipitationProbability, period.WindSpeed))
		p.periods.AddItem(card, 0, 1, false)
	}
}

func (p *Panel) dailyText(daily []weatherprobe.Daily) string {
	var builder strings.Builder
	builder.WriteString("[turquoise]7-DAY OUTLOOK[-]\n")
	for _, forecast := range daily {
		visual := weatherprobe.ConditionVisual(forecast.WeatherCode)
		builder.WriteString(fmt.Sprintf("[green]%-6s[-] [%s]%-13s[-] [green]%2.0f/%2.0f°C[-]  [blue]%3.0f%%[-]  [gray]%2.0f km/h[-]\n",
			forecast.Date.Format("Mon 02"), visual.Color, visual.Label, forecast.High, forecast.Low, forecast.PrecipitationProbability, forecast.WindSpeed))
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

func (p *Panel) selectedLocation() string {
	if p.active < 0 || p.active >= len(p.forecasts) {
		return ""
	}
	return p.forecasts[p.active].Location.Name
}

func indexForLocation(forecasts []weatherprobe.Forecast, location string) int {
	for index, forecast := range forecasts {
		if forecast.Location.Name == location {
			return index
		}
	}
	return 0
}

func cardView(title string) *tview.TextView {
	card := textView()
	theme.Box(card.Box, title)
	return card
}

func textView() *tview.TextView {
	return tview.NewTextView().SetDynamicColors(true).SetWrap(false)
}
