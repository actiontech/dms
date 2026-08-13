package biz

import (
	"testing"
	"time"
)

func TestCalcActiveMinutes(t *testing.T) {
	base := time.Date(2026, 6, 22, 9, 0, 0, 0, time.Local)
	tests := []struct {
		name  string
		times []time.Time
		want  int
	}{
		{
			name:  "single event",
			times: []time.Time{base},
			want:  0,
		},
		{
			name: "continuous session",
			times: []time.Time{
				base,
				base.Add(10 * time.Minute),
				base.Add(20 * time.Minute),
			},
			want: 20,
		},
		{
			name: "two sessions with gap",
			times: []time.Time{
				base,
				base.Add(10 * time.Minute),
				base.Add(50 * time.Minute),
				base.Add(60 * time.Minute),
			},
			want: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcActiveMinutes(tt.times, 30)
			if got != tt.want {
				t.Fatalf("CalcActiveMinutes() = %d, want %d", got, tt.want)
			}
		})
	}
}
