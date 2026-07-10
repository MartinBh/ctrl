package weather

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
	"github.com/martinbhatta/ctrl/internal/theme"
)

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
	theme.Box(p.root.Box, "WEATHER")
	p.SetLoading()

	return p
}

func (p *Panel) Primitive() tview.Primitive {
	return p.root
}

func (p *Panel) SetLoading() {
	p.forecasts = nil
	p.status.SetText("[gray]Open-Meteo · resolving your public-IP location[-]")
	p.cards.Clear()
	card := cardView("LOCAL WEATHER")
	card.SetText("[gray]Resolving location and loading current conditions...[-]")
	p.cards.AddItem(card, 0, 1, false)
	p.detailTitle.SetText("[gray]Forecast detail will appear after loading.[-]")
	p.periods.Clear()
	p.daily.SetText("")
}

func (p *Panel) SetForecasts(forecasts []weatherprobe.Forecast) {
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

	p.status.SetText(p.statusText())
	p.render()
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
	forecast := p.forecasts[0]
	if forecast.Err != nil {
		p.detailTitle.SetText(fmt.Sprintf("[red]%s forecast unavailable: %s[-]", displayLocationName(forecast.Location.Name), tview.Escape(forecast.Err.Error())))
		p.periods.Clear()
		p.daily.SetText("")
		return
	}

	p.detailTitle.SetText(fmt.Sprintf("[turquoise]%s forecast[-]  [gray]location resolved from your public IP[-]", displayLocationName(forecast.Location.Name)))
	p.renderPeriods(forecast.Hourly)
	p.daily.SetText(p.dailyText(forecast.Daily))
}

func (p *Panel) statusText() string {
	stale := make([]string, 0, len(p.forecasts))
	unavailable := make([]string, 0, len(p.forecasts))
	for _, forecast := range p.forecasts {
		switch {
		case p.stale[forecast.Location.Name]:
			stale = append(stale, forecast.Location.Name)
		case forecast.Err != nil:
			unavailable = append(unavailable, forecast.Location.Name)
		}
	}

	switch {
	case len(p.forecasts) == 0:
		return "[red]Open-Meteo · weather is unavailable[-]"
	case len(unavailable) > 0:
		return fmt.Sprintf("[red]Open-Meteo · weather unavailable for %s[-]", displayLocationNames(unavailable))
	case len(stale) > 0:
		return fmt.Sprintf("[yellow]Open-Meteo · showing last successful conditions for %s[-]", displayLocationNames(stale))
	default:
		return "[gray]Open-Meteo · current conditions[-]"
	}
}

func displayLocationName(name string) string {
	return tview.Escape(name)
}

func displayLocationNames(names []string) string {
	escaped := make([]string, len(names))
	for index, name := range names {
		escaped[index] = displayLocationName(name)
	}
	return strings.Join(escaped, ", ")
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

func cardView(title string) *tview.TextView {
	card := textView()
	theme.Box(card.Box, title)
	return card
}

func textView() *tview.TextView {
	return tview.NewTextView().SetDynamicColors(true).SetWrap(false)
}
