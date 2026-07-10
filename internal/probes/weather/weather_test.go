package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestForecastsBuildsRequestAndParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got, want := query.Get("timezone"), "Asia/Seoul"; got != want {
			t.Fatalf("timezone = %q, want %q", got, want)
		}
		if got, want := query.Get("forecast_days"), "7"; got != want {
			t.Fatalf("forecast_days = %q, want %q", got, want)
		}
		for _, field := range []string{"temperature_2m", "weather_code", "wind_speed_10m"} {
			if !strings.Contains(query.Get("current")+query.Get("hourly"), field) {
				t.Fatalf("weather request missing %q", field)
			}
		}
		if got := query.Get("latitude"); got != "37.517235" && got != "37.600020" {
			t.Fatalf("unexpected latitude = %q", got)
		}
		_, _ = w.Write([]byte(sampleResponse()))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL
	client.httpClient = server.Client()
	client.now = func() time.Time { return mustTime(t, "2026-07-10T10:20") }

	forecasts := client.Forecasts(context.Background())
	if len(forecasts) != 2 {
		t.Fatalf("forecast count = %d, want 2", len(forecasts))
	}
	first := forecasts[0]
	if first.Err != nil {
		t.Fatalf("Forecasts() error = %v", first.Err)
	}
	if got, want := first.Current.Temperature, 26.5; got != want {
		t.Fatalf("current temperature = %v, want %v", got, want)
	}
	if got, want := len(first.Hourly), 8; got != want {
		t.Fatalf("hourly count = %d, want %d", got, want)
	}
	if got, want := first.Hourly[0].Time.Format("15:04"), "11:00"; got != want {
		t.Fatalf("first hourly time = %q, want %q", got, want)
	}
	if got, want := len(first.Daily), 7; got != want {
		t.Fatalf("daily count = %d, want %d", got, want)
	}
}

func TestForecastsKeepsLocationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL
	client.httpClient = server.Client()

	forecasts := client.Forecasts(context.Background())
	if forecasts[0].Err == nil {
		t.Fatal("Forecasts() error = nil, want request error")
	}
	if got, want := ErrorSummary(forecasts), "Gangnam-gu, Sangbong-dong"; got != want {
		t.Fatalf("ErrorSummary() = %q, want %q", got, want)
	}
}

func TestForecastsKeepsMalformedResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"current":`))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL
	client.httpClient = server.Client()

	forecasts := client.Forecasts(context.Background())
	if forecasts[0].Err == nil || !strings.Contains(forecasts[0].Err.Error(), "decode weather response") {
		t.Fatalf("Forecasts() error = %v, want decode error", forecasts[0].Err)
	}
}

func TestHourlyConditionsRejectsMismatchedFields(t *testing.T) {
	response := apiResponse{}
	response.Hourly.Time = []string{"2026-07-10T11:00"}
	response.Hourly.Temperature = []float64{26}
	response.Hourly.ApparentTemperature = []float64{27}
	response.Hourly.PrecipitationProbability = []float64{20}
	response.Hourly.WeatherCode = []int{1}

	_, err := response.hourlyConditions(mustTime(t, "2026-07-10T10:00"))
	if err == nil || !strings.Contains(err.Error(), "invalid hourly weather response") {
		t.Fatalf("hourlyConditions() error = %v, want mismatched-field error", err)
	}
}

func TestRequestURLContainsLocation(t *testing.T) {
	client := NewClient()
	requestURL, err := url.Parse(client.requestURL(Locations[1]))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := requestURL.Query().Get("longitude"), "127.092830"; got != want {
		t.Fatalf("longitude = %q, want %q", got, want)
	}
}

func TestConditionAndWindDirection(t *testing.T) {
	if got, want := Condition(95), "Thunderstorm"; got != want {
		t.Fatalf("Condition(95) = %q, want %q", got, want)
	}
	if got, want := WindDirection(360), "N"; got != want {
		t.Fatalf("WindDirection(360) = %q, want %q", got, want)
	}
	if got, want := WindDirection(-90), "W"; got != want {
		t.Fatalf("WindDirection(-90) = %q, want %q", got, want)
	}
}

func TestConditionVisualUsesTerminalSafeGlyphs(t *testing.T) {
	tests := []struct {
		code  int
		label string
		color string
	}{
		{code: 0, label: "Clear", color: "yellow"},
		{code: 3, label: "Overcast", color: "gray"},
		{code: 61, label: "Rain", color: "blue"},
		{code: 71, label: "Snow", color: "white"},
		{code: 95, label: "Thunderstorm", color: "red"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			visual := ConditionVisual(tt.code)
			if visual.Label != tt.label || visual.Color != tt.color {
				t.Fatalf("ConditionVisual(%d) = %#v, want label %q and color %q", tt.code, visual, tt.label, tt.color)
			}
			if len(visual.Glyph) != 3 {
				t.Fatalf("glyph line count = %d, want 3", len(visual.Glyph))
			}
		})
	}
}

func TestSummarizePeriods(t *testing.T) {
	periods := SummarizePeriods([]Hourly{
		{Time: mustTime(t, "2026-07-10T10:00"), Temperature: 24, PrecipitationProbability: 20, WeatherCode: 2, WindSpeed: 4},
		{Time: mustTime(t, "2026-07-10T11:00"), Temperature: 26, PrecipitationProbability: 80, WeatherCode: 61, WindSpeed: 7},
		{Time: mustTime(t, "2026-07-10T13:00"), Temperature: 28, PrecipitationProbability: 10, WeatherCode: 1, WindSpeed: 8},
		{Time: mustTime(t, "2026-07-10T16:00"), Temperature: 27, PrecipitationProbability: 40, WeatherCode: 3, WindSpeed: 5},
	})

	if got, want := len(periods), 2; got != want {
		t.Fatalf("period count = %d, want %d", got, want)
	}
	if got, want := periods[0], (Period{Label: "Morning", Condition: 61, Low: 24, High: 26, PrecipitationProbability: 80, WindSpeed: 7}); got != want {
		t.Fatalf("morning period = %#v, want %#v", got, want)
	}
	if got, want := periods[1], (Period{Label: "Afternoon", Condition: 3, Low: 27, High: 28, PrecipitationProbability: 40, WindSpeed: 8}); got != want {
		t.Fatalf("afternoon period = %#v, want %#v", got, want)
	}
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := parseTime(value)
	if err != nil {
		t.Fatalf("parseTime(%q) error = %v", value, err)
	}
	return parsed
}

func sampleResponse() string {
	return `{
  "current": {"time":"2026-07-10T10:00","temperature_2m":26.5,"apparent_temperature":27.1,"relative_humidity_2m":71,"precipitation":0,"weather_code":1,"wind_speed_10m":8.3,"wind_direction_10m":90},
  "hourly": {
    "time":["2026-07-10T09:00","2026-07-10T10:00","2026-07-10T11:00","2026-07-10T12:00","2026-07-10T13:00","2026-07-10T14:00","2026-07-10T15:00","2026-07-10T16:00","2026-07-10T17:00","2026-07-10T18:00","2026-07-10T19:00"],
    "temperature_2m":[20,21,22,23,24,25,26,27,28,29,30],
    "apparent_temperature":[20,21,22,23,24,25,26,27,28,29,30],
    "precipitation_probability":[0,1,2,3,4,5,6,7,8,9,10],
    "weather_code":[0,0,1,1,2,2,3,3,61,61,61],
    "wind_speed_10m":[1,2,3,4,5,6,7,8,9,10,11]
  },
  "daily": {
    "time":["2026-07-10","2026-07-11","2026-07-12","2026-07-13","2026-07-14","2026-07-15","2026-07-16"],
    "weather_code":[0,1,2,3,61,80,95],
    "temperature_2m_max":[30,31,32,33,34,35,36],
    "temperature_2m_min":[20,21,22,23,24,25,26],
    "precipitation_probability_max":[0,1,2,3,4,5,6],
    "wind_speed_10m_max":[1,2,3,4,5,6,7]
  }
}`
}
