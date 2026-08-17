package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/mipmip/skull2/internal/config"
)

// Repo is the provider-agnostic repository model shared by the cache, sync and
// TUI. See BRIEFING.md §7.
type Repo struct {
	// Owner is the owning user or organization.
	Owner string
	// Name is the repository name.
	Name string
	// Description is the repository description.
	Description string
	// SSHURL is the git SSH clone URL.
	SSHURL string
	// HTTPSURL is the git HTTPS clone URL.
	HTTPSURL string
	// WebURL is the browser URL for the repository.
	WebURL string
	// DefaultBranch is the repository's default branch.
	DefaultBranch string
	// Archived reports whether the repository is archived.
	Archived bool
	// Fork reports whether the repository is a fork.
	Fork bool
	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time
}

// Provider lists repositories for a set of owners.
type Provider interface {
	// ListRepos returns the repositories for the given owners. An empty owners
	// slice means all accessible repositories.
	ListRepos(ctx context.Context, owners []string) ([]Repo, error)
	// ListOwners discovers the organizations/groups the authenticated user
	// belongs to, mapped to the same owner strings ListRepos accepts. It does
	// NOT include the user's own account (that is the SelfOwner sentinel added
	// by config.ResolveOwners).
	ListOwners(ctx context.Context) ([]string, error)
}

// ZeroDiscoveryWarning returns a human-readable warning when all_owners is
// enabled but discovery returned no organizations, likely a missing token scope
// (e.g. GitHub read:org). It returns "" when a warning is not warranted so
// callers can `if w := ...; w != "" { log }`.
func ZeroDiscoveryWarning(p *config.Provider, discovered []string) string {
	if !p.AllOwners || len(discovered) > 0 {
		return ""
	}
	return fmt.Sprintf(
		"provider %q: all_owners is enabled but discovery found no organizations; "+
			"the token may be missing an org-read scope (e.g. GitHub read:org)",
		p.Name,
	)
}

// FilterRepos drops archived and/or fork repositories according to the
// provider's include_archived and include_forks configuration flags. Nil flags
// use the documented defaults (archived excluded, forks included).
func FilterRepos(p *config.Provider, repos []Repo) []Repo {
	includeArchived := false
	if p.IncludeArchived != nil {
		includeArchived = *p.IncludeArchived
	}
	includeForks := true
	if p.IncludeForks != nil {
		includeForks = *p.IncludeForks
	}

	out := make([]Repo, 0, len(repos))
	for _, r := range repos {
		if r.Archived && !includeArchived {
			continue
		}
		if r.Fork && !includeForks {
			continue
		}
		out = append(out, r)
	}
	return out
}
