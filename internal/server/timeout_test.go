package server

import "testing"

func TestClampTaskTimeout(t *testing.T) {
	cases := []struct {
		name      string
		requested int
		max       int
		want      int
	}{
		{"no max keeps requested", 300, 0, 300},
		{"no max keeps unlimited", 0, 0, 0},
		{"under max kept", 300, 1800, 300},
		{"equal max kept", 1800, 1800, 1800},
		{"over max capped", 3600, 1800, 1800},
		{"unlimited capped to max", 0, 1800, 1800},
		{"negative request capped to max", -5, 1800, 1800},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := clampTaskTimeout(c.requested, c.max); got != c.want {
				t.Fatalf("clampTaskTimeout(%d, %d) = %d, want %d", c.requested, c.max, got, c.want)
			}
		})
	}
}
