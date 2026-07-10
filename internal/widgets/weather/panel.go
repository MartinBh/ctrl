package weather

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
	"github.com/martinbhatta/ctrl/internal/theme"
)

type Panel struct {
	table          *tview.Table
	lastSuccessful map[string]weatherprobe.Forecast
}

func NewPanel() *Panel {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	theme.Box(table.Box, "WEATHER")

	return &Panel{table: table, lastSuccessful: make(map[string]weatherprobe.Forecast)}
}

func (p *Panel) Primitive() tview.Primitive {
	return p.table
}

func (p *Panel) SetLoading() {
	p.table.Clear()
	p.table.SetCell(0, 0, titleCell("Weather: Open-Meteo"))
	p.table.SetCell(1, 0, mutedCell("Loading Gangnam-gu and Sangbong-dong forecasts..."))
}

func (p *Panel) SetForecasts(forecasts []weatherprobe.Forecast) {
	p.table.Clear()
	row := 0
	p.table.SetCell(row, 0, titleCell("Weather: Open-Meteo"))
	row++

	for _, forecast := range forecasts {
		stale := false
		if forecast.Err != nil {
			if previous, ok := p.lastSuccessful[forecast.Location.Name]; ok {
				forecast = previous
				stale = true
			}
		} else {
			p.lastSuccessful[forecast.Location.Name] = forecast
		}

		p.table.SetCell(row, 0, titleCell(forecast.Location.Name))
		row++

		if stale {
			p.table.SetCell(row, 0, errorCell("Stale"))
			p.table.SetCell(row, 1, mutedCell("Latest weather request failed; showing the last forecast."))
			row++
		}

		if forecast.Err != nil && !stale {
			p.table.SetCell(row, 0, errorCell("Unavailable"))
			p.table.SetCell(row, 1, mutedCell(forecast.Err.Error()))
			row += 2
			continue
		}

		p.table.SetCell(row, 0, headerCell("Now"))
		p.table.SetCell(row, 1, accentCell(weatherprobe.Condition(forecast.Current.WeatherCode)))
		p.table.SetCell(row, 2, mutedCell(fmt.Sprintf("%.0f°C feels %.0f°C", forecast.Current.Temperature, forecast.Current.ApparentTemperature)))
		p.table.SetCell(row, 3, mutedCell(fmt.Sprintf("%d%% humidity", int(forecast.Current.Humidity))))
		p.table.SetCell(row, 4, mutedCell(fmt.Sprintf("%.0f%% precip", forecast.Current.Precipitation)))
		p.table.SetCell(row, 5, mutedCell(fmt.Sprintf("%.0f km/h %s", forecast.Current.WindSpeed, weatherprobe.WindDirection(forecast.Current.WindDirection))))
		row++

		p.table.SetCell(row, 0, headerCell("Next 8 hours"))
		p.table.SetCell(row, 1, headerCell("Condition"))
		p.table.SetCell(row, 2, headerCell("Temp"))
		p.table.SetCell(row, 3, headerCell("Precip"))
		p.table.SetCell(row, 4, headerCell("Wind"))
		row++
		for _, hourly := range forecast.Hourly {
			p.table.SetCell(row, 0, accentCell(hourly.Time.Format("15:04")))
			p.table.SetCell(row, 1, mutedCell(weatherprobe.Condition(hourly.WeatherCode)))
			p.table.SetCell(row, 2, mutedCell(fmt.Sprintf("%.0f°C", hourly.Temperature)))
			p.table.SetCell(row, 3, mutedCell(fmt.Sprintf("%.0f%%", hourly.PrecipitationProbability)))
			p.table.SetCell(row, 4, mutedCell(fmt.Sprintf("%.0f km/h", hourly.WindSpeed)))
			row++
		}

		p.table.SetCell(row, 0, headerCell("7-day forecast"))
		p.table.SetCell(row, 1, headerCell("Condition"))
		p.table.SetCell(row, 2, headerCell("High / Low"))
		p.table.SetCell(row, 3, headerCell("Precip"))
		p.table.SetCell(row, 4, headerCell("Wind"))
		row++
		for _, daily := range forecast.Daily {
			p.table.SetCell(row, 0, accentCell(daily.Date.Format("Mon 02")))
			p.table.SetCell(row, 1, mutedCell(weatherprobe.Condition(daily.WeatherCode)))
			p.table.SetCell(row, 2, mutedCell(fmt.Sprintf("%.0f° / %.0f°", daily.High, daily.Low)))
			p.table.SetCell(row, 3, mutedCell(fmt.Sprintf("%.0f%%", daily.PrecipitationProbability)))
			p.table.SetCell(row, 4, mutedCell(fmt.Sprintf("%.0f km/h", daily.WindSpeed)))
			row++
		}
		row++
	}
}

func titleCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).SetTextColor(theme.ColorPrimary)
}

func headerCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(theme.ColorPrimary).
		SetSelectable(false)
}

func accentCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).SetTextColor(theme.ColorAccent)
}

func mutedCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).SetTextColor(theme.ColorMuted)
}

func errorCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).SetTextColor(tcell.ColorRed)
}
