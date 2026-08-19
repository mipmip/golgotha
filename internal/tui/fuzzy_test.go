package tui

import "testing"

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		name   string
		target string
		query  string
		want   bool
	}{
		{"empty query matches", "mipmip/huphop", "", true},
		{"exact substring", "mipmip/huphop", "huph", true},
		{"subsequence", "mipmip/huphop", "mhp", true},
		{"case insensitive", "MipMip/Huphop", "miphu", true},
		{"owner and name subsequence", "technative-b-v/foo", "tnfoo", true},
		{"not a subsequence", "mipmip/huphop", "zzz", false},
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

func TestCompileQuery(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		strategy string
		target   string
		want     bool
	}{
		// Empty query matches everything.
		{"empty matches", "", searchFuzzy, "mipmip/huphop", true},
		{"whitespace only matches", "   ", searchFuzzy, "mipmip/huphop", true},

		// Bare term follows the configured strategy.
		{"fuzzy bare subsequence", "nvs", searchFuzzy, "mip/nix-services", true},
		{"substring bare rejects subsequence", "nivis", searchSubstring, "mip/nix-services", false},
		{"substring bare accepts literal", "nivis", searchSubstring, "org/nivis-app", true},

		// The ' prefix toggles to the opposite strategy.
		{"quote forces substring in fuzzy mode", "'nivis", searchFuzzy, "mip/nix-services", false},
		{"quote forces substring literal hit", "'nivis", searchFuzzy, "org/nivis", true},
		{"quote forces fuzzy in substring mode", "'nvs", searchSubstring, "mip/nix-services", true},

		// Prefix / suffix anchors.
		{"prefix hit", "^mip", searchFuzzy, "mipmip/huphop", true},
		{"prefix miss", "^huph", searchFuzzy, "mipmip/huphop", false},
		{"suffix hit", "hop$", searchFuzzy, "mipmip/huphop", true},
		{"suffix miss", "mip$", searchFuzzy, "mipmip/huphop", false},
		{"equal hit", "^mipmip/huphop$", searchSubstring, "mipmip/huphop", true},
		{"equal miss", "^huphop$", searchSubstring, "mipmip/huphop", false},

		// Negation.
		{"negate excludes match", "!huph", searchSubstring, "mipmip/huphop", false},
		{"negate keeps non-match", "!zzz", searchSubstring, "mipmip/huphop", true},
		{"negate prefix", "!^mip", searchFuzzy, "mipmip/huphop", false},

		// AND across terms.
		{"and both present", "mip hop", searchSubstring, "mipmip/huphop", true},
		{"and one missing", "mip zzz", searchSubstring, "mipmip/huphop", false},
		{"and with negation", "mip !fork", searchSubstring, "mipmip/huphop", true},

		// OR within an AND group: "a b | c" = a AND (b OR c).
		{"or right branch", "mip zzz | hop", searchSubstring, "mipmip/huphop", true},
		{"or left branch", "mip hop | zzz", searchSubstring, "mipmip/huphop", true},
		{"or neither branch", "mip zzz | yyy", searchSubstring, "mipmip/huphop", false},

		// Case-insensitivity.
		{"case insensitive substring", "HUPH", searchSubstring, "MipMip/Huphop", true},

		// Dangling / lone operators are ignored gracefully.
		{"leading or ignored", "| hop", searchSubstring, "mipmip/huphop", true},
		{"lone quote ignored", "'", searchFuzzy, "mipmip/huphop", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := compileQuery(tc.query, tc.strategy)
			if got := m.match(tc.target); got != tc.want {
				t.Fatalf("compileQuery(%q, %q).match(%q) = %v, want %v",
					tc.query, tc.strategy, tc.target, got, tc.want)
			}
		})
	}
}
