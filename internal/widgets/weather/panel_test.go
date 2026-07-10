package weather

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rivo/tview"

	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
)

func TestPanelShowsLoadingCards(t *testing.T) {
	panel := NewPanel()

	if got, want := panel.cards.GetItemCount(), 1; got != want {
		t.Fatalf("loading card count = %d, want %d", got, want)
	}
	if got := panel.status.GetText(true); !strings.Contains(got, "Open-Meteo") {
		t.Fatalf("loading status = %q, want attribution", got)
	}
}

func TestPanelShowsLocalForecast(t *testing.T) {
	panel := NewPanel()
	panel.SetForecasts([]weatherprobe.Forecast{forecast(t, "New York", 1)})

	card := panel.cards.GetItem(0).(*tview.TextView).GetText(true)
	if !strings.Contains(card, "PARTLY CLOUDY") || !strings.Contains(card, "25°C") {
		t.Fatalf("card = %q, want condition and temperature", card)
	}
	if got := panel.detailTitle.GetText(true); !strings.Contains(got, "New York forecast") {
		t.Fatalf("detail title = %q, want local forecast", got)
	}
	if got := panel.status.GetText(true); !strings.Contains(got, "current conditions") {
		t.Fatalf("status = %q, want current conditions", got)
	}
	if got := panel.daily.GetText(true); !strings.Contains(got, "7-DAY OUTLOOK") {
		t.Fatalf("daily outlook = %q, want heading", got)
	}
}

func TestPanelRetainsLastForecastWhenRefreshFails(t *testing.T) {
	panel := NewPanel()
	success := forecast(t, "New York", 0)
	panel.SetForecasts([]weatherprobe.Forecast{success})
	panel.SetForecasts([]weatherprobe.Forecast{{Location: success.Location, Err: errors.New("request timed out")}})

	card := panel.cards.GetItem(0).(*tview.TextView).GetText(true)
	if !strings.Contains(card, "STALE") || !strings.Contains(card, "CLEAR") {
		t.Fatalf("stale card = %q, want stale retained forecast", card)
	}
	if got := panel.status.GetText(true); !strings.Contains(got, "last successful") {
		t.Fatalf("status = %q, want stale status", got)
	}
}

func TestPanelShowsUnavailableStatus(t *testing.T) {
	panel := NewPanel()
	panel.SetForecasts([]weatherprobe.Forecast{{Location: weatherprobe.Location{Name: "Local weather"}, Err: errors.New("request timed out")}})

	if got := panel.status.GetText(true); !strings.Contains(got, "weather unavailable") {
		t.Fatalf("status = %q, want unavailable status", got)
	}
}

func forecast(t *testing.T, name string, weatherCode int) weatherprobe.Forecast {
	t.Helper()
	return weatherprobe.Forecast{
		Location: weatherprobe.Location{Name: name},
		Current:  weatherprobe.Current{Temperature: 25, ApparentTemperature: 27, Humidity: 70, Precipitation: 0.1, WeatherCode: weatherCode, WindSpeed: 10, WindDirection: 90},
		Hourly: []weatherprobe.Hourly{
			{Time: localTime(t, "2026-07-10T10:00"), Temperature: 25, WeatherCode: weatherCode, PrecipitationProbability: 60, WindSpeed: 10},
			{Time: localTime(t, "2026-07-10T13:00"), Temperature: 27, WeatherCode: weatherCode, PrecipitationProbability: 80, WindSpeed: 11},
		},
		Daily: []weatherprobe.Daily{{Date: localTime(t, "2026-07-10T00:00"), High: 28, Low: 23, WeatherCode: weatherCode, PrecipitationProbability: 80, WindSpeed: 12}},
	}
}

func localTime(t *testing.T, value string) time.Time {
	t.Helper()
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	parsed, err := time.ParseInLocation("2006-01-02T15:04", value, location)
	if err != nil {
		t.Fatalf("ParseInLocation() error = %v", err)
	}
	return parsed
}
