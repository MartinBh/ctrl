package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const forecastURL = "https://api.open-meteo.com/v1/forecast"

var Locations = []Location{
	{Name: "Gangnam-gu", Latitude: 37.517235, Longitude: 127.047325},
	{Name: "Sangbong-dong", Latitude: 37.60002, Longitude: 127.09283},
}

type Location struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type Current struct {
	Time                time.Time
	Temperature         float64
	ApparentTemperature float64
	Humidity            float64
	Precipitation       float64
	WeatherCode         int
	WindSpeed           float64
	WindDirection       float64
}

type Hourly struct {
	Time                     time.Time
	Temperature              float64
	ApparentTemperature      float64
	PrecipitationProbability float64
	WeatherCode              int
	WindSpeed                float64
}

type Daily struct {
	Date                     time.Time
	WeatherCode              int
	High                     float64
	Low                      float64
	PrecipitationProbability float64
	WindSpeed                float64
}

type Forecast struct {
	Location Location
	Current  Current
	Hourly   []Hourly
	Daily    []Daily
	Err      error
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	now        func() time.Time
}

func NewClient() *Client {
	return &Client{
		baseURL:    forecastURL,
		httpClient: http.DefaultClient,
		now:        time.Now,
	}
}

func (c *Client) Forecasts(ctx context.Context) []Forecast {
	forecasts := make([]Forecast, len(Locations))
	var waitGroup sync.WaitGroup
	for index, location := range Locations {
		waitGroup.Add(1)
		go func(index int, location Location) {
			defer waitGroup.Done()
			forecasts[index] = c.forecast(ctx, location)
		}(index, location)
	}
	waitGroup.Wait()
	return forecasts
}

func (c *Client) forecast(ctx context.Context, location Location) Forecast {
	forecast := Forecast{Location: location}
	requestURL := c.requestURL(location)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		forecast.Err = err
		return forecast
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		forecast.Err = err
		return forecast
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		forecast.Err = fmt.Errorf("weather request returned %s", response.Status)
		return forecast
	}

	var payload apiResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		forecast.Err = fmt.Errorf("decode weather response: %w", err)
		return forecast
	}

	forecast.Current, err = payload.currentCondition()
	if err != nil {
		forecast.Err = err
		return forecast
	}
	forecast.Hourly, err = payload.hourlyConditions(c.now())
	if err != nil {
		forecast.Err = err
		return forecast
	}
	forecast.Daily, err = payload.dailyConditions()
	if err != nil {
		forecast.Err = err
	}
	return forecast
}

func (c *Client) requestURL(location Location) string {
	values := url.Values{}
	values.Set("latitude", fmt.Sprintf("%.6f", location.Latitude))
	values.Set("longitude", fmt.Sprintf("%.6f", location.Longitude))
	values.Set("timezone", "Asia/Seoul")
	values.Set("forecast_days", "7")
	values.Set("current", "temperature_2m,apparent_temperature,relative_humidity_2m,precipitation,weather_code,wind_speed_10m,wind_direction_10m")
	values.Set("hourly", "temperature_2m,apparent_temperature,precipitation_probability,weather_code,wind_speed_10m")
	values.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max")
	return c.baseURL + "?" + values.Encode()
}

type apiResponse struct {
	Current struct {
		Time                string  `json:"time"`
		Temperature         float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Humidity            float64 `json:"relative_humidity_2m"`
		Precipitation       float64 `json:"precipitation"`
		WeatherCode         int     `json:"weather_code"`
		WindSpeed           float64 `json:"wind_speed_10m"`
		WindDirection       float64 `json:"wind_direction_10m"`
	} `json:"current"`
	Hourly struct {
		Time                     []string  `json:"time"`
		Temperature              []float64 `json:"temperature_2m"`
		ApparentTemperature      []float64 `json:"apparent_temperature"`
		PrecipitationProbability []float64 `json:"precipitation_probability"`
		WeatherCode              []int     `json:"weather_code"`
		WindSpeed                []float64 `json:"wind_speed_10m"`
	} `json:"hourly"`
	Daily struct {
		Time                     []string  `json:"time"`
		WeatherCode              []int     `json:"weather_code"`
		High                     []float64 `json:"temperature_2m_max"`
		Low                      []float64 `json:"temperature_2m_min"`
		PrecipitationProbability []float64 `json:"precipitation_probability_max"`
		WindSpeed                []float64 `json:"wind_speed_10m_max"`
	} `json:"daily"`
}

func (r apiResponse) currentCondition() (Current, error) {
	timestamp, err := parseTime(r.Current.Time)
	if err != nil {
		return Current{}, fmt.Errorf("parse current time: %w", err)
	}

	return Current{Time: timestamp, Temperature: r.Current.Temperature, ApparentTemperature: r.Current.ApparentTemperature, Humidity: r.Current.Humidity, Precipitation: r.Current.Precipitation, WeatherCode: r.Current.WeatherCode, WindSpeed: r.Current.WindSpeed, WindDirection: r.Current.WindDirection}, nil
}

func (r apiResponse) hourlyConditions(now time.Time) ([]Hourly, error) {
	length := len(r.Hourly.Time)
	if err := matchingLengths(length, len(r.Hourly.Temperature), len(r.Hourly.ApparentTemperature), len(r.Hourly.PrecipitationProbability), len(r.Hourly.WeatherCode), len(r.Hourly.WindSpeed)); err != nil {
		return nil, fmt.Errorf("invalid hourly weather response: %w", err)
	}

	hourly := make([]Hourly, 0, length)
	for index, rawTime := range r.Hourly.Time {
		timestamp, err := parseTime(rawTime)
		if err != nil {
			return nil, fmt.Errorf("parse hourly time: %w", err)
		}
		if timestamp.Before(now.In(timestamp.Location()).Truncate(time.Hour).Add(time.Hour)) {
			continue
		}
		hourly = append(hourly, Hourly{Time: timestamp, Temperature: r.Hourly.Temperature[index], ApparentTemperature: r.Hourly.ApparentTemperature[index], PrecipitationProbability: r.Hourly.PrecipitationProbability[index], WeatherCode: r.Hourly.WeatherCode[index], WindSpeed: r.Hourly.WindSpeed[index]})
	}

	sort.Slice(hourly, func(i int, j int) bool { return hourly[i].Time.Before(hourly[j].Time) })
	if len(hourly) > 8 {
		hourly = hourly[:8]
	}
	return hourly, nil
}

func (r apiResponse) dailyConditions() ([]Daily, error) {
	length := len(r.Daily.Time)
	if err := matchingLengths(length, len(r.Daily.WeatherCode), len(r.Daily.High), len(r.Daily.Low), len(r.Daily.PrecipitationProbability), len(r.Daily.WindSpeed)); err != nil {
		return nil, fmt.Errorf("invalid daily weather response: %w", err)
	}

	daily := make([]Daily, length)
	for index, rawDate := range r.Daily.Time {
		date, err := parseDate(rawDate)
		if err != nil {
			return nil, fmt.Errorf("parse daily date: %w", err)
		}
		daily[index] = Daily{Date: date, WeatherCode: r.Daily.WeatherCode[index], High: r.Daily.High[index], Low: r.Daily.Low[index], PrecipitationProbability: r.Daily.PrecipitationProbability[index], WindSpeed: r.Daily.WindSpeed[index]}
	}
	return daily, nil
}

func matchingLengths(length int, values ...int) error {
	for _, value := range values {
		if value != length {
			return fmt.Errorf("time has %d values, related field has %d", length, value)
		}
	}
	return nil
}

func parseTime(value string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("2006-01-02T15:04", value, location)
}

func parseDate(value string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("2006-01-02", value, location)
}

func Condition(code int) string {
	switch code {
	case 0:
		return "Clear"
	case 1, 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45, 48:
		return "Fog"
	case 51, 53, 55, 56, 57:
		return "Drizzle"
	case 61, 63, 65, 66, 67:
		return "Rain"
	case 71, 73, 75, 77:
		return "Snow"
	case 80, 81, 82:
		return "Showers"
	case 85, 86:
		return "Snow showers"
	case 95, 96, 99:
		return "Thunderstorm"
	default:
		return "Unknown"
	}
}

func WindDirection(degrees float64) string {
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	index := int((degrees+22.5)/45) % len(directions)
	return directions[index]
}

func ErrorSummary(forecasts []Forecast) string {
	failed := make([]string, 0, len(forecasts))
	for _, forecast := range forecasts {
		if forecast.Err != nil {
			failed = append(failed, forecast.Location.Name)
		}
	}
	return strings.Join(failed, ", ")
}
