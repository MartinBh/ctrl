package battery

import (
	"testing"

	batteryprobe "github.com/martinbhatta/ctrl/internal/probes/battery"
)

func TestStatusTextPresent(t *testing.T) {
	got := statusText(batteryprobe.Status{
		Present: true,
		Percent: 82.4,
		State:   "charging",
	})

	if got != "82% charging" {
		t.Fatalf("statusText = %q, want 82%% charging", got)
	}
}

func TestStatusTextMissing(t *testing.T) {
	got := statusText(batteryprobe.Status{
		Present: false,
		State:   "no battery detected",
	})

	if got != "no battery detected" {
		t.Fatalf("statusText = %q, want no battery detected", got)
	}
}
