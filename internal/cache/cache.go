package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mipmip/huphop/internal/provider"
)

// OwnerState records one owner in the cache's owner index and whether its
// repositories have been fetched. A nil FetchedAt means the owner has been
// discovered but its repositories have not been fetched yet.
type OwnerState struct {
	// Name is the owner/org name (as used by provider ListRepos).
	Name string `json:"name"`
	// FetchedAt is when this owner's repos were fetched; nil = discovered-only.
	FetchedAt *time.Time `json:"fetched_at"`
}

// Cache is the persisted per-provider repository snapshot (v2). It keeps the
// flat Repos list (each Repo carries its Owner) and adds an owner index with
// per-owner fetch state so consumers can distinguish "discovered" from
// "fetched". FetchedAt is retained for backward compatibility and reflects the
// most recent whole-provider fetch/discovery time.
type Cache struct {
	// FetchedAt is the time the repositories were last fetched from the provider.
	// Retained for backward compatibility with the legacy flat shape.
	FetchedAt time.Time `json:"fetched_at"`
	// DiscoveredAt is when the owner index was last (re)discovered.
	DiscoveredAt time.Time `json:"discovered_at,omitempty"`
	// Owners is the owner index with per-owner fetch state. May be empty for a
	// legacy cache loaded from the old flat shape (see UnmarshalJSON).
	Owners []OwnerState `json:"owners,omitempty"`
	// Repos is the list of repositories for the provider.
	Repos []provider.Repo `json:"repos"`
}

// legacyCache is the old flat shape used only for backward-compatible reads: a
// cache written before the owner index existed had no "owners" key.
type legacyCache struct {
	FetchedAt    time.Time       `json:"fetched_at"`
	DiscoveredAt time.Time       `json:"discovered_at"`
	Owners       []OwnerState    `json:"owners"`
	Repos        []provider.Repo `json:"repos"`
}

// UnmarshalJSON reads both the v2 shape and the legacy flat shape. When the file
// has no owner index, every owner present in Repos is treated as already fetched
// (with FetchedAt as its timestamp), so old caches load transparently.
func (c *Cache) UnmarshalJSON(data []byte) error {
	var raw legacyCache
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.FetchedAt = raw.FetchedAt
	c.DiscoveredAt = raw.DiscoveredAt
	c.Owners = raw.Owners
	c.Repos = raw.Repos

	if len(c.Owners) == 0 && len(c.Repos) > 0 {
		// Legacy flat cache: synthesize a fully-fetched owner index from repos.
		ts := raw.FetchedAt
		seen := make(map[string]struct{})
		for _, r := range c.Repos {
			if _, dup := seen[r.Owner]; dup {
				continue
			}
			seen[r.Owner] = struct{}{}
			t := ts
			c.Owners = append(c.Owners, OwnerState{Name: r.Owner, FetchedAt: &t})
		}
	}
	return nil
}

// SetOwners replaces the owner index with the discovered owner names, preserving
// existing per-owner fetch timestamps for owners that are still present, and
// records DiscoveredAt. Owners no longer present are dropped along with their
// repos. Passing the config.SelfOwner sentinel ("") is supported.
func (c *Cache) SetOwners(discoveredAt time.Time, names []string) {
	prev := make(map[string]*time.Time, len(c.Owners))
	for _, o := range c.Owners {
		prev[o.Name] = o.FetchedAt
	}
	keep := make(map[string]struct{}, len(names))
	owners := make([]OwnerState, 0, len(names))
	for _, n := range names {
		if _, dup := keep[n]; dup {
			continue
		}
		keep[n] = struct{}{}
		owners = append(owners, OwnerState{Name: n, FetchedAt: prev[n]})
	}
	c.Owners = owners
	c.DiscoveredAt = discoveredAt

	// Drop repos whose owner is no longer in the index.
	if len(c.Repos) > 0 {
		repos := c.Repos[:0:0]
		for _, r := range c.Repos {
			if _, ok := keep[r.Owner]; ok {
				repos = append(repos, r)
			}
		}
		c.Repos = repos
	}
}

// MarkOwnerFetched replaces the given owner's repositories with repos and marks
// it fetched at fetchedAt, adding the owner to the index if it was not present.
// Other owners' repos and state are preserved. FetchedAt is also advanced.
func (c *Cache) MarkOwnerFetched(owner string, repos []provider.Repo, fetchedAt time.Time) {
	// Rebuild Repos: drop this owner's repos, keep the rest, append the new ones.
	kept := make([]provider.Repo, 0, len(c.Repos)+len(repos))
	for _, r := range c.Repos {
		if r.Owner != owner {
			kept = append(kept, r)
		}
	}
	kept = append(kept, repos...)
	c.Repos = kept
	c.FetchedAt = fetchedAt

	ts := fetchedAt
	for i := range c.Owners {
		if c.Owners[i].Name == owner {
			c.Owners[i].FetchedAt = &ts
			return
		}
	}
	c.Owners = append(c.Owners, OwnerState{Name: owner, FetchedAt: &ts})
}

// UnfetchedOwners returns the names of owners that have been discovered but whose
// repositories have not been fetched (FetchedAt == nil), in index order.
func (c *Cache) UnfetchedOwners() []string {
	var out []string
	for _, o := range c.Owners {
		if o.FetchedAt == nil {
			out = append(out, o.Name)
		}
	}
	return out
}

// OwnerFetched reports whether the named owner exists in the index and has been
// fetched.
func (c *Cache) OwnerFetched(owner string) bool {
	for _, o := range c.Owners {
		if o.Name == owner {
			return o.FetchedAt != nil
		}
	}
	return false
}

// ReposFor returns the cached repositories for the named owner.
func (c *Cache) ReposFor(owner string) []provider.Repo {
	var out []provider.Repo
	for _, r := range c.Repos {
		if r.Owner == owner {
			out = append(out, r)
		}
	}
	return out
}

// OwnerNames returns the owner names in index order.
func (c *Cache) OwnerNames() []string {
	out := make([]string, 0, len(c.Owners))
	for _, o := range c.Owners {
		out = append(out, o.Name)
	}
	return out
}

// Dir returns the huphop cache directory, honoring $XDG_CACHE_HOME and falling
// back to ~/.cache/huphop.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "huphop"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".cache", "huphop"), nil
}

// Path returns the cache file path for a provider (<dir>/<provider>.json).
func Path(providerName string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, providerName+".json"), nil
}

// Save writes the cache for providerName atomically: it marshals to JSON, writes
// a temporary file in the cache directory, then renames it into place so a
// concurrent reader never observes a partial file.
func Save(providerName string, c Cache) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating cache dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache for %q: %w", providerName, err)
	}

	tmp, err := os.CreateTemp(dir, providerName+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp cache file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we fail before the rename.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp cache file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp cache file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp cache file: %w", err)
	}

	final := filepath.Join(dir, providerName+".json")
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("renaming cache file into place: %w", err)
	}
	return nil
}

// Load reads and unmarshals the cache for providerName. A missing cache file
// returns an error wrapping os.ErrNotExist; callers that tolerate absence should
// use LoadOrEmpty.
func Load(providerName string) (Cache, error) {
	path, err := Path(providerName)
	if err != nil {
		return Cache{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Cache{}, err
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return Cache{}, fmt.Errorf("parsing cache %s: %w", path, err)
	}
	return c, nil
}

// LoadOrEmpty loads the cache for providerName, tolerating a missing file. The
// returned bool reports whether a cache file existed. A missing file yields an
// empty Cache, ok=false and a nil error; other errors are returned as-is.
func LoadOrEmpty(providerName string) (Cache, bool, error) {
	c, err := Load(providerName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Cache{}, false, nil
		}
		return Cache{}, false, err
	}
	return c, true, nil
}
