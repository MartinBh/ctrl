package usage

import (
	"testing"

	"github.com/martinbhatta/ctrl/internal/probes"
)

func TestLevelForUsedPct(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want probes.Level
	}{
		{name: "normal", pct: 74.9, want: probes.LevelOK},
		{name: "warning threshold", pct: 75, want: probes.LevelWarning},
		{name: "warning", pct: 89.9, want: probes.LevelWarning},
		{name: "error threshold", pct: 90, want: probes.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := levelForUsedPct(tt.pct); got != tt.want {
				t.Fatalf("levelForUsedPct(%v) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}
