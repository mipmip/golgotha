package tui

import (
	"testing"

	"github.com/mipmip/golgotha/internal/provider"
)

func TestTriStateMatch(t *testing.T) {
	cases := []struct {
		state    triState
		v        bool
		expected bool
	}{
		{triAll, true, true},
		{triAll, false, true},
		{triOnly, true, true},
		{triOnly, false, false},
		{triHide, true, false},
		{triHide, false, true},
	}
	for _, c := range cases {
		if got := c.state.match(c.v); got != c.expected {
			t.Errorf("triState(%d).match(%v) = %v, want %v", c.state, c.v, got, c.expected)
		}
	}
}

func TestTriStateCycle(t *testing.T) {
	if triAll.cycle() != triOnly || triOnly.cycle() != triHide || triHide.cycle() != triAll {
		t.Fatal("tri-state cycle order wrong")
	}
}

func TestVisFacetMatchAndCycle(t *testing.T) {
	if visAll.cycle() != visPublic || visPublic.cycle() != visPrivate ||
		visPrivate.cycle() != visInternal || visInternal.cycle() != visAll {
		t.Fatal("visibility cycle order wrong")
	}
	if !visAll.match("private") {
		t.Error("visAll should match anything")
	}
	if !visPrivate.match(provider.VisibilityPrivate) || visPrivate.match(provider.VisibilityPublic) {
		t.Error("visPrivate match wrong")
	}
	if !visInternal.match(provider.VisibilityInternal) {
		t.Error("visInternal match wrong")
	}
}

func TestFacetsMatchAndCompose(t *testing.T) {
	pubRepo := provider.Repo{Name: "a", Visibility: "public"}
	privFork := provider.Repo{Name: "b", Visibility: "private", Fork: true}
	archived := provider.Repo{Name: "c", Visibility: "public", Archived: true}
	unknownVis := provider.Repo{Name: "d"} // empty -> normalizes to public

	// No active facets: everything matches.
	f := facets{}
	if f.active() {
		t.Fatal("default facets should be inactive")
	}
	for _, r := range []provider.Repo{pubRepo, privFork, archived, unknownVis} {
		if !f.match(r) {
			t.Fatalf("inactive facets should match %s", r.Name)
		}
	}

	// fork:hide drops the fork.
	f = facets{fork: triHide}
	if !f.active() || f.match(privFork) || !f.match(pubRepo) {
		t.Fatal("fork:hide predicate wrong")
	}

	// archived:only keeps only archived.
	f = facets{archived: triOnly}
	if f.match(pubRepo) || !f.match(archived) {
		t.Fatal("archived:only predicate wrong")
	}

	// vis:public treats unknown/empty as public.
	f = facets{vis: visPublic}
	if !f.match(unknownVis) || f.match(privFork) {
		t.Fatal("vis:public normalization wrong")
	}

	// Compose: vis:private AND fork:only -> only privFork.
	f = facets{vis: visPrivate, fork: triOnly}
	if !f.match(privFork) || f.match(pubRepo) || f.match(archived) {
		t.Fatal("composed facet predicate wrong")
	}
}

func TestFacetsStatus(t *testing.T) {
	if s := (facets{}).status(); s != "" {
		t.Fatalf("inactive status should be empty, got %q", s)
	}
	f := facets{fork: triHide, vis: visPrivate}
	if s := f.status(); s != "fork:hide vis:private" {
		t.Fatalf("status = %q", s)
	}
	f = facets{archived: triOnly}
	if s := f.status(); s != "archived:only" {
		t.Fatalf("status = %q", s)
	}
}
