package usage

import "testing"

func TestProgressBar(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{name: "empty", pct: 0, want: "[----------]"},
		{name: "half", pct: 50, want: "[#####-----]"},
		{name: "full", pct: 100, want: "[##########]"},
		{name: "clamps low", pct: -10, want: "[----------]"},
		{name: "clamps high", pct: 150, want: "[##########]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := progressBar(tt.pct, 10); got != tt.want {
				t.Fatalf("progressBar(%v, 10) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}
