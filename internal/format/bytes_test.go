package format

import "testing"

func TestBytes(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
		want  string
	}{
		{name: "bytes", value: 512, want: "512 B"},
		{name: "megabytes", value: 512 * 1024 * 1024, want: "512 MB"},
		{name: "single digit gigabytes", value: 1536 * 1024 * 1024, want: "1.5 GB"},
		{name: "double digit gigabytes", value: 12 * 1024 * 1024 * 1024, want: "12 GB"},
		{name: "terabytes", value: 1 * 1024 * 1024 * 1024 * 1024, want: "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Bytes(tt.value); got != tt.want {
				t.Fatalf("Bytes(%d) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
