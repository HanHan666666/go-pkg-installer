package ui

import "testing"

func TestScrollReachedEnd(t *testing.T) {
	tests := []struct {
		name   string
		yview  string
		expect bool
	}{
		{
			name:   "end",
			yview:  "0.0 1.0",
			expect: true,
		},
		{
			name:   "not-end",
			yview:  "0.0 0.5",
			expect: false,
		},
		{
			name:   "invalid",
			yview:  "invalid",
			expect: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scrollReachedEnd(tc.yview); got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}
