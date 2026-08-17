package tui

// filters.go holds the interactive facet-filter state and the pure filtering
// predicate. Facets narrow the currently cached repositories (Model A,
// narrow-only) and compose (AND) with the fuzzy query. State is in-memory for
// the session and starts at "all" so behavior is unchanged until a facet is
// touched.

import (
	"strings"

	"github.com/mipmip/skull2/internal/provider"
)

// triState is a three-way facet: include everything, only matches, or hide
// matches. It applies to the fork and archived booleans.
type triState int

const (
	triAll  triState = iota // no constraint
	triOnly                 // keep only rows where the attribute is true
	triHide                 // drop rows where the attribute is true
)

// cycle advances the tri-state all -> only -> hide -> all.
func (t triState) cycle() triState { return (t + 1) % 3 }

// match reports whether a row with the given attribute value passes this facet.
func (t triState) match(v bool) bool {
	switch t {
	case triOnly:
		return v
	case triHide:
		return !v
	default:
		return true
	}
}

// label renders the facet value for the status line ("all"/"only"/"hide").
func (t triState) label() string {
	switch t {
	case triOnly:
		return "only"
	case triHide:
		return "hide"
	default:
		return "all"
	}
}

// visFacet is the value-cycle visibility facet: all, then each visibility value.
type visFacet int

const (
	visAll visFacet = iota
	visPublic
	visPrivate
	visInternal
)

// cycle advances all -> public -> private -> internal -> all.
func (v visFacet) cycle() visFacet { return (v + 1) % 4 }

// match reports whether a repo's visibility string passes this facet.
func (v visFacet) match(vis string) bool {
	switch v {
	case visPublic:
		return vis == provider.VisibilityPublic
	case visPrivate:
		return vis == provider.VisibilityPrivate
	case visInternal:
		return vis == provider.VisibilityInternal
	default:
		return true
	}
}

// label renders the facet value for the status line.
func (v visFacet) label() string {
	switch v {
	case visPublic:
		return provider.VisibilityPublic
	case visPrivate:
		return provider.VisibilityPrivate
	case visInternal:
		return provider.VisibilityInternal
	default:
		return "all"
	}
}

// facets is the interactive filter state applied to the repo list.
type facets struct {
	fork     triState
	archived triState
	vis      visFacet
}

// active reports whether any facet is set away from "all".
func (f facets) active() bool {
	return f.fork != triAll || f.archived != triAll || f.vis != visAll
}

// match reports whether a repo passes every active facet (AND).
func (f facets) match(r provider.Repo) bool {
	return f.fork.match(r.Fork) &&
		f.archived.match(r.Archived) &&
		f.vis.match(provider.NormalizeVisibility(r.Visibility))
}

// status renders the active facets for the status line, e.g. "fork:hide
// vis:private". It returns "" when no facet is active.
func (f facets) status() string {
	var parts []string
	if f.fork != triAll {
		parts = append(parts, "fork:"+f.fork.label())
	}
	if f.archived != triAll {
		parts = append(parts, "archived:"+f.archived.label())
	}
	if f.vis != visAll {
		parts = append(parts, "vis:"+f.vis.label())
	}
	return strings.Join(parts, " ")
}
