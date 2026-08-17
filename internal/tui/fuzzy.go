package tui

import "strings"

// fuzzyMatch reports whether every rune of the (lowercased) query appears in
// target in order, as a subsequence. An empty query matches everything. The
// match is case-insensitive.
func fuzzyMatch(target, query string) bool {
	if query == "" {
		return true
	}
	target = strings.ToLower(target)
	query = strings.ToLower(query)

	ti := 0
	for _, q := range query {
		found := false
		for ti < len(target) {
			r := rune(target[ti])
			ti++
			if r == q {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
