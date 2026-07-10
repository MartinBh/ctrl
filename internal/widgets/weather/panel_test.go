package weather

import (
	"errors"
	"testing"
	"time"

	weatherprobe "github.com/martinbhatta/ctrl/internal/probes/weather"
)

func TestPanelShowsLoadingState(t *testing.T) {
	panel := NewPanel()
	panel.SetLoading()

	if got, want := panel.table.GetCell(0, 0).Text, "Weather: Open-Meteo"; got != want {
		t.Fatalf("attribution = %q, want %q", got, want)
	}
	if got := panel.table.GetCell(1, 0).Text; got == "" {
		t.Fatal("loading text is empty")
	}
}

func TestPanelShowsForecastAndLocationError(t *testing.T) {
	panel := NewPanel()
	panel.SetForecasts([]weatherprobe.Forecast{
		{
			Location: weatherprobe.Location{Name: "Gangnam-gu"},
			Current:  weatherprobe.Current{Temperature: 25, ApparentTemperature: 26, Humidity: 70, WeatherCode: 1, WindSpeed: 10, WindDirection: 90},
			Hourly:   []weatherprobe.Hourly{{Time: seoulTime(t, "2026-07-10T11:00"), Temperature: 26, WeatherCode: 2, WindSpeed: 11}},
			Daily:    []weatherprobe.Daily{{Date: seoulTime(t, "2026-07-10T00:00"), High: 30, Low: 21, WeatherCode: 3, WindSpeed: 12}},
		},
		{Location: weatherprobe.Location{Name: "Sangbong-dong"}, Err: errors.New("request timed out")},
	})

	if got, want := panel.table.GetCell(2, 1).Text, "Partly cloudy"; got != want {
		t.Fatalf("current condition = %q, want %q", got, want)
	}
	if got, want := panel.table.GetCell(4, 0).Text, "11:00"; got != want {
		t.Fatalf("hourly time = %q, want %q", got, want)
	}
	if got, want := panel.table.GetCell(6, 0).Text, "Fri 10"; got != want {
		t.Fatalf("daily date = %q, want %q", got, want)
	}
	if got, want := panel.table.GetCell(9, 0).Text, "Unavailable"; got != want {
		t.Fatalf("error state = %q, want %q", got, want)
	}
}

func TestPanelRetainsLastForecastWhenRefreshFails(t *testing.T) {
	panel := NewPanel()
	success := weatherprobe.Forecast{
		Location: weatherprobe.Location{Name: "Gangnam-gu"},
		Current:  weatherprobe.Current{WeatherCode: 0},
	}
	panel.SetForecasts([]weatherprobe.Forecast{success})
	panel.SetForecasts([]weatherprobe.Forecast{{Location: success.Location, Err: errors.New("request timed out")}})

	if got, want := panel.table.GetCell(2, 0).Text, "Stale"; got != want {
		t.Fatalf("stale state = %q, want %q", got, want)
	}
	if got, want := panel.table.GetCell(3, 1).Text, "Clear"; got != want {
		t.Fatalf("retained condition = %q, want %q", got, want)
	}
}

func seoulTime(t *testing.T, value string) time.Time {
	t.Helper()
	location, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	parsed, err := time.ParseInLocation("2006-01-02T15:04", value, location)
	if err != nil {
		t.Fatalf("ParseInLocation() error = %v", err)
	}
	return parsed
}
