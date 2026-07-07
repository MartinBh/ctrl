package battery

import (
	"testing"

	batterylib "github.com/distatus/battery"

	"github.com/martinbhatta/ctrl/internal/probes"
)

func TestSummarizeNoBattery(t *testing.T) {
	got := Summarize(nil)

	if got.Present {
		t.Fatal("Summarize(nil) marked battery present")
	}
	if got.State != "no battery detected" {
		t.Fatalf("State = %q, want no battery detected", got.State)
	}
	if got.Level != probes.LevelMuted {
		t.Fatalf("Level = %q, want %q", got.Level, probes.LevelMuted)
	}
}

func TestSummarizeChargingBattery(t *testing.T) {
	got := Summarize([]*batterylib.Battery{{
		State:   batterylib.State{Raw: batterylib.Charging},
		Current: 80,
		Full:    100,
	}})

	if !got.Present {
		t.Fatal("Summarize marked battery missing")
	}
	if got.Percent != 80 {
		t.Fatalf("Percent = %v, want 80", got.Percent)
	}
	if got.State != "charging" {
		t.Fatalf("State = %q, want charging", got.State)
	}
	if got.Level != probes.LevelOK {
		t.Fatalf("Level = %q, want %q", got.Level, probes.LevelOK)
	}
}

func TestSummarizeLowDischargingBattery(t *testing.T) {
	got := Summarize([]*batterylib.Battery{{
		State:   batterylib.State{Raw: batterylib.Discharging},
		Current: 10,
		Full:    100,
	}})

	if got.Level != probes.LevelError {
		t.Fatalf("Level = %q, want %q", got.Level, probes.LevelError)
	}
}

func TestSummarizeMultipleBatteries(t *testing.T) {
	got := Summarize([]*batterylib.Battery{
		{State: batterylib.State{Raw: batterylib.Full}, Current: 40, Full: 50},
		{State: batterylib.State{Raw: batterylib.Full}, Current: 50, Full: 50},
	})

	if got.Percent != 90 {
		t.Fatalf("Percent = %v, want 90", got.Percent)
	}
	if got.Detail != "2 batteries detected" {
		t.Fatalf("Detail = %q, want 2 batteries detected", got.Detail)
	}
}

func TestLevelForState(t *testing.T) {
	tests := []struct {
		name  string
		state batterylib.AgnosticState
		pct   float64
		want  probes.Level
	}{
		{name: "charging", state: batterylib.Charging, pct: 10, want: probes.LevelOK},
		{name: "discharging ok", state: batterylib.Discharging, pct: 31, want: probes.LevelOK},
		{name: "discharging warning", state: batterylib.Discharging, pct: 30, want: probes.LevelWarning},
		{name: "discharging error", state: batterylib.Discharging, pct: 14.9, want: probes.LevelError},
		{name: "unknown", state: batterylib.Unknown, pct: 80, want: probes.LevelWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := levelForState(tt.state, tt.pct); got != tt.want {
				t.Fatalf("levelForState(%v, %v) = %q, want %q", tt.state, tt.pct, got, tt.want)
			}
		})
	}
}
