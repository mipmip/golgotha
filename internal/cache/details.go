package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mipmip/huphop/internal/provider"
)

// Details is the persisted per-repository detail snapshot, cached separately
// from the lean per-provider list cache. It stores the lazily-fetched tier-2
// metadata (stars, topics, language) and the RAW README markdown; rendering to
// styled terminal text happens at view time so width/theme adapt. See openspec
// add-repo-details.
type Details struct {
	// FetchedAt is when these details were fetched from the provider.
	FetchedAt time.Time `json:"fetched_at"`
	// Stars is the repository's star count.
	Stars int `json:"stars"`
	// Topics is the repository's topics/tags.
	Topics []string `json:"topics"`
	// Language is the repository's primary language.
	Language string `json:"language"`
	// ReadmeMarkdown is the RAW README markdown (never pre-rendered).
	ReadmeMarkdown string `json:"readme_md"`
}

// DetailsDir returns the directory holding per-repository detail caches for a
// provider: <cache-dir>/details/<provider>.
func DetailsDir(providerName string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "details", providerName), nil
}

// detailFileName maps an owner/name to the on-disk file name using the
// "<owner>__<repo>.json" convention, sanitizing path separators so nested
// owners (GitLab groups) stay a single flat file.
func detailFileName(owner, name string) string {
	return sanitizeSegment(owner) + "__" + sanitizeSegment(name) + ".json"
}

// sanitizeSegment replaces filesystem-unsafe characters (path separators) in an
// owner or repo name so it is safe as a single flat file-name segment.
func sanitizeSegment(s string) string {
	r := strings.NewReplacer("/", "-", string(os.PathSeparator), "-", "..", "-")
	return r.Replace(s)
}

// DetailsPath returns the detail cache file path for a repository:
// <cache-dir>/details/<provider>/<owner>__<repo>.json.
func DetailsPath(providerName, owner, name string) (string, error) {
	dir, err := DetailsDir(providerName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, detailFileName(owner, name)), nil
}

// SaveDetails writes a repository's detail cache atomically, creating the
// provider detail directory as needed. It is separate from the list cache and
// never touches <provider>.json.
func SaveDetails(providerName, owner, name string, d Details) error {
	dir, err := DetailsDir(providerName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating detail cache dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling details for %q %s/%s: %w", providerName, owner, name, err)
	}

	tmp, err := os.CreateTemp(dir, "detail.*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp detail file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp detail file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp detail file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp detail file: %w", err)
	}

	final := filepath.Join(dir, detailFileName(owner, name))
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("renaming detail file into place: %w", err)
	}
	return nil
}

// LoadDetails reads a repository's detail cache. A missing file returns an error
// wrapping os.ErrNotExist; callers that tolerate absence should use
// LoadDetailsOrEmpty.
func LoadDetails(providerName, owner, name string) (Details, error) {
	path, err := DetailsPath(providerName, owner, name)
	if err != nil {
		return Details{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Details{}, err
	}
	var d Details
	if err := json.Unmarshal(data, &d); err != nil {
		return Details{}, fmt.Errorf("parsing detail cache %s: %w", path, err)
	}
	return d, nil
}

// LoadDetailsOrEmpty loads a repository's detail cache, tolerating a missing
// file. The returned bool reports whether a cache file existed. A missing file
// yields a zero Details, ok=false and a nil error; other errors are returned.
func LoadDetailsOrEmpty(providerName, owner, name string) (Details, bool, error) {
	d, err := LoadDetails(providerName, owner, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Details{}, false, nil
		}
		return Details{}, false, err
	}
	return d, true, nil
}

// RefreshDetails builds a Details snapshot from freshly-fetched provider details
// and raw README markdown, stamps it with the fetch time and persists it,
// returning the stored value. It is the write path used by a manual refresh (or
// first open) so the on-disk cache and the returned value stay identical.
func RefreshDetails(providerName, owner, name string, pd provider.Details, readmeMD string, fetchedAt time.Time) (Details, error) {
	d := Details{
		FetchedAt:      fetchedAt,
		Stars:          pd.Stars,
		Topics:         pd.Topics,
		Language:       pd.Language,
		ReadmeMarkdown: readmeMD,
	}
	if err := SaveDetails(providerName, owner, name, d); err != nil {
		return Details{}, err
	}
	return d, nil
}
