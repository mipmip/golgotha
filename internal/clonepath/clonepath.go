package clonepath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mipmip/huphop/internal/config"
)

// Data holds the fields available to a clone-path template. See BRIEFING.md §5.
type Data struct {
	// BaseDir is the configured base dir (tilde-expanded, absolute).
	BaseDir string
	// Provider is the provider name.
	Provider string
	// Type is the provider type (github, codeberg, gitlab).
	Type string
	// Short is the provider short code.
	Short string
	// Host is the provider host.
	Host string
	// Owner is the owner/org in upstream casing.
	Owner string
	// OwnerLower is the lowercased owner/org.
	OwnerLower string
	// Repo is the repo name in upstream casing.
	Repo string
	// RepoLower is the lowercased repo name.
	RepoLower string
}

// NewData builds template Data for a repo under the given provider, deriving
// the lowercase variants automatically.
func NewData(baseDir string, p *config.Provider, host, owner, repo string) Data {
	return Data{
		BaseDir:    baseDir,
		Provider:   p.Name,
		Type:       string(p.Type),
		Short:      p.Short,
		Host:       host,
		Owner:      owner,
		OwnerLower: strings.ToLower(owner),
		Repo:       repo,
		RepoLower:  strings.ToLower(repo),
	}
}

// TemplateFor returns the effective clone-path template for a provider: its
// per-provider override when set, otherwise the global template.
func TemplateFor(cfg *config.Config, p *config.Provider) string {
	if p.ClonePatternTpl != "" {
		return p.ClonePatternTpl
	}
	return cfg.ClonePatternTpl
}

// Render renders the given template with data and returns a cleaned, absolute
// path guaranteed to be within data.BaseDir. It rejects templates that escape
// the base directory and reports missing template keys as errors.
func Render(tpl string, data Data) (string, error) {
	t, err := template.New("clonepath").Option("missingkey=error").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parsing clone_pattern_tpl: %w", err)
	}

	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("rendering clone_pattern_tpl: %w", err)
	}

	rendered, err := expandHome(sb.String())
	if err != nil {
		return "", err
	}
	rendered = filepath.Clean(rendered)

	if !filepath.IsAbs(rendered) {
		abs, err := filepath.Abs(rendered)
		if err != nil {
			return "", fmt.Errorf("resolving absolute path: %w", err)
		}
		rendered = abs
	}

	base := filepath.Clean(data.BaseDir)
	if !withinBase(base, rendered) {
		return "", fmt.Errorf("clone path %q escapes base_dir %q", rendered, base)
	}
	return rendered, nil
}

// RenderFor is a convenience wrapper that resolves the effective template for a
// provider and renders it for the given repo.
func RenderFor(cfg *config.Config, p *config.Provider, host, owner, repo string) (string, error) {
	data := NewData(cfg.BaseDir, p, host, owner, repo)
	return Render(TemplateFor(cfg, p), data)
}

// withinBase reports whether target is base itself or a descendant of base.
func withinBase(base, target string) bool {
	if target == base {
		return true
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// expandHome expands a leading ~ or ~/ to the user's home directory.
func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[len("~/"):]), nil
	}
	return path, nil
}
