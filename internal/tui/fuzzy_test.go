package tui

import "testing"

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		name   string
		target string
		query  string
		want   bool
	}{
		{"empty query matches", "mipmip/golgotha", "", true},
		{"exact substring", "mipmip/golgotha", "golg", true},
		{"subsequence", "mipmip/golgotha", "mgl", true},
		{"case insensitive", "MipMip/Golgotha", "mipgo", true},
		{"owner and name subsequence", "technative-b-v/foo", "tnfoo", true},
		{"not a subsequence", "mipmip/golgotha", "zzz", false},
		{"out of order fails", "abc", "cba", false},
		{"longer than target", "ab", "abc", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := fuzzyMatch(tc.target, tc.query); got != tc.want {
				t.Fatalf("fuzzyMatch(%q, %q) = %v, want %v", tc.target, tc.query, got, tc.want)
			}
		})
	}
}
