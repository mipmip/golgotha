package tui

import "testing"

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		name   string
		target string
		query  string
		want   bool
	}{
		{"empty query matches", "mipmip/skull2", "", true},
		{"exact substring", "mipmip/skull2", "skull", true},
		{"subsequence", "mipmip/skull2", "msk", true},
		{"case insensitive", "MipMip/Skull2", "mipsk", true},
		{"owner and name subsequence", "technative-b-v/foo", "tnfoo", true},
		{"not a subsequence", "mipmip/skull2", "zzz", false},
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
