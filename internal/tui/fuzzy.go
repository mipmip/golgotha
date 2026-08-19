package tui

import "strings"

// Search strategies for a bare (unanchored) filter term. They mirror the
// config values config.SearchFuzzy / config.SearchSubstring; kept as local
// constants so the tui package does not import config just for the strings.
const (
	searchFuzzy     = "fuzzy"
	searchSubstring = "substring"
)

// matchKind is how a single token is compared against a target string.
type matchKind int

const (
	kindFuzzy     matchKind = iota // subsequence
	kindSubstring                  // contains
	kindPrefix                     // starts-with (^)
	kindSuffix                     // ends-with ($)
	kindEqual                      // exact equality (^…$)
)

// token is one atom of the filter query: a lowercased text plus how it matches
// and whether the result is negated (leading !).
type token struct {
	kind   matchKind
	negate bool
	text   string
}

// match reports whether target (already lowercased) satisfies the token.
func (t token) match(target string) bool {
	var ok bool
	switch t.kind {
	case kindSubstring:
		ok = strings.Contains(target, t.text)
	case kindPrefix:
		ok = strings.HasPrefix(target, t.text)
	case kindSuffix:
		ok = strings.HasSuffix(target, t.text)
	case kindEqual:
		ok = target == t.text
	default: // kindFuzzy
		ok = subseqMatch(target, t.text)
	}
	if t.negate {
		return !ok
	}
	return ok
}

// matcher is a compiled query: a conjunction (AND) of OR-groups. A target
// matches when every group has at least one matching token.
type matcher struct {
	groups [][]token
}

// match reports whether target satisfies the compiled query. An empty matcher
// (empty query) matches everything.
func (m matcher) match(target string) bool {
	if len(m.groups) == 0 {
		return true
	}
	target = strings.ToLower(target)
	for _, group := range m.groups {
		if !orGroupMatch(group, target) {
			return false
		}
	}
	return true
}

func orGroupMatch(group []token, target string) bool {
	for _, tok := range group {
		if tok.match(target) {
			return true
		}
	}
	return false
}

// compileQuery parses an fzf-style extended-search query into a matcher.
//
// Grammar: whitespace-separated terms are ANDed; a bare "|" between terms folds
// the adjacent terms into a single OR-group ("a b | c" = a AND (b OR c)). Each
// term may be negated with a leading "!", anchored with a leading "^" (prefix)
// and/or trailing "$" (suffix), or toggled to the opposite of the configured
// strategy with a leading "'". A bare term uses strategy (fuzzy or substring).
func compileQuery(query, strategy string) matcher {
	fields := strings.Fields(query)
	var groups [][]token
	orPending := false
	for _, f := range fields {
		if f == "|" {
			// A dangling "|" (no current group) is ignored; otherwise the next
			// term joins the current OR-group.
			if len(groups) > 0 {
				orPending = true
			}
			continue
		}
		tok, ok := parseToken(f, strategy)
		if !ok {
			continue
		}
		if orPending {
			last := len(groups) - 1
			groups[last] = append(groups[last], tok)
			orPending = false
			continue
		}
		groups = append(groups, []token{tok})
	}
	return matcher{groups: groups}
}

// parseToken parses one whitespace-delimited atom. It returns ok=false when the
// atom carries no matchable text (e.g. a lone operator), so the caller skips it.
func parseToken(atom, strategy string) (token, bool) {
	tok := token{}
	if strings.HasPrefix(atom, "!") {
		tok.negate = true
		atom = atom[1:]
	}

	switch {
	case strings.HasPrefix(atom, "'"):
		// Quote toggles to the opposite of the configured bare strategy.
		atom = atom[1:]
		if strategy == searchSubstring {
			tok.kind = kindFuzzy
		} else {
			tok.kind = kindSubstring
		}
	default:
		prefix := strings.HasPrefix(atom, "^")
		suffix := strings.HasSuffix(atom, "$")
		if prefix {
			atom = atom[1:]
		}
		if suffix {
			atom = atom[:len(atom)-1]
		}
		switch {
		case prefix && suffix:
			tok.kind = kindEqual
		case prefix:
			tok.kind = kindPrefix
		case suffix:
			tok.kind = kindSuffix
		case strategy == searchSubstring:
			tok.kind = kindSubstring
		default:
			tok.kind = kindFuzzy
		}
	}

	if atom == "" {
		return token{}, false
	}
	tok.text = strings.ToLower(atom)
	return tok, true
}

// subseqMatch reports whether every rune of the (lowercased) query appears in
// target in order, as a subsequence. An empty query matches everything. The
// match is case-insensitive; target is assumed already lowercased by the caller
// but is lowered defensively for standalone use.
func subseqMatch(target, query string) bool {
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

// fuzzyMatch reports whether query is a subsequence of target (case-insensitive).
// Retained as the fuzzy primitive used by the compiled matcher and existing
// tests.
func fuzzyMatch(target, query string) bool {
	return subseqMatch(target, query)
}
