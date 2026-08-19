package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/mipmip/huphop/internal/config"
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
	// Visibility is the repository visibility: "public", "private" or
	// "internal". Unknown/empty values normalize to "public".
	Visibility string
	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time
}

// Repository visibility values used by Repo.Visibility.
const (
	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	VisibilityInternal = "internal"
)

// NormalizeVisibility maps a raw provider visibility string to one of the
// recognized values. Anything unrecognized (including empty) normalizes to
// "public".
func NormalizeVisibility(v string) string {
	switch v {
	case VisibilityPublic, VisibilityPrivate, VisibilityInternal:
		return v
	default:
		return VisibilityPublic
	}
}

// visibilityFromPrivate maps a boolean "private" flag (GitHub, Gitea/Codeberg)
// to the visibility string.
func visibilityFromPrivate(private bool) string {
	if private {
		return VisibilityPrivate
	}
	return VisibilityPublic
}

// Details holds a repository's lazily-fetched extended metadata (tier-2): the
// star count, topics and primary language. It is fetched on demand and cached
// apart from the lean list cache. See openspec add-repo-details.
type Details struct {
	// Stars is the repository's stargazer/star count.
	Stars int
	// Topics is the list of repository topics/tags.
	Topics []string
	// Language is the repository's primary language ("" when unknown).
	Language string
}

// Provider lists repositories for a set of owners.
type Provider interface {
	// ListRepos returns the repositories for the given owners. An empty owners
	// slice means all accessible repositories.
	ListRepos(ctx context.Context, owners []string) ([]Repo, error)
	// ListOwners discovers the organizations/groups the authenticated user
	// belongs to, mapped to the same owner strings ListRepos accepts. It does
	// NOT include the user's own account (that is the configured Username, added
	// by config.ResolveOwners).
	ListOwners(ctx context.Context) ([]string, error)
	// RepoDetails fetches a single repository's extended details (stars, topics,
	// primary language). owner is the repository's owner/namespace and name is
	// the repository name.
	RepoDetails(ctx context.Context, owner, name string) (Details, error)
	// Readme fetches a single repository's raw README markdown. A repository with
	// no README yields an empty string and a nil error (a clean "not found",
	// never a hard error), so callers can degrade gracefully.
	Readme(ctx context.Context, owner, name string) (string, error)
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
