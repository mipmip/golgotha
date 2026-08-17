package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mipmip/skull2/internal/provider"
)

// Cache is the persisted per-provider repository snapshot.
type Cache struct {
	// FetchedAt is the time the repositories were fetched from the provider.
	FetchedAt time.Time `json:"fetched_at"`
	// Repos is the list of repositories for the provider.
	Repos []provider.Repo `json:"repos"`
}

// Dir returns the skull2 cache directory, honoring $XDG_CACHE_HOME and falling
// back to ~/.cache/skull2.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "skull2"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".cache", "skull2"), nil
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
