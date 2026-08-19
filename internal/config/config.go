package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProviderType identifies a supported git-hosting provider kind.
type ProviderType string

// Supported provider types.
const (
	ProviderGitHub   ProviderType = "github"
	ProviderCodeberg ProviderType = "codeberg"
	ProviderGitLab   ProviderType = "gitlab"
)

// DefaultClonePatternTpl is the default clone-path Go text/template applied when
// neither the global config nor a provider overrides it. See BRIEFING.md §5.
const DefaultClonePatternTpl = "{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}"

// Clone protocols.
const (
	ProtocolSSH   = "ssh"
	ProtocolHTTPS = "https"
)

// Config is the top-level huphop configuration parsed from config.yaml.
type Config struct {
	// BaseDir is the root under which all trees are created. Tilde-expanded to
	// an absolute path after loading.
	BaseDir string `yaml:"base_dir"`
	// ClonePatternTpl is the global default clone-path template.
	ClonePatternTpl string `yaml:"clone_pattern_tpl"`
	// DefaultMode is the TUI mode used when --mode is not given. Defaults to
	// "management".
	DefaultMode string `yaml:"default_mode"`
	// Modes maps a mode name to its configuration (chrome + mode settings). When
	// omitted, a built-in "management" mode reproducing the standard layout is
	// synthesized.
	Modes map[string]ModeConfig `yaml:"modes"`
	// Providers is the configured list of git providers.
	Providers []Provider `yaml:"providers"`
}

// TUI modes and the placeable chrome-element vocabulary.
const (
	// ModeManagement is the default browsing mode (Enter opens repo details).
	ModeManagement = "management"
	// ModeMultiplex clones-if-needed then runs a switch_command (tmux, etc.).
	ModeMultiplex = "multiplex"
)

// KnownElements is the set of placeable header/footer element names. The repo
// list is the implicit body and is never named here.
var KnownElements = map[string]bool{
	"breadcrumb":         true,
	"action_menu":        true,
	"filter":             true,
	"facet_status":       true,
	"status_message":     true,
	"position_indicator": true,
	"switch_hint":        true,
	"clone_status":       true,
}

// ModeConfig is a single TUI mode: ordered header/footer element slots plus any
// mode-specific settings.
type ModeConfig struct {
	// Header lists the elements rendered above the repo list, in order.
	Header []string `yaml:"header"`
	// Footer lists the elements rendered below the repo list, in order.
	Footer []string `yaml:"footer"`
	// SwitchCommand is the per-repo command template run by the multiplex mode's
	// primary action (rendered then executed without a shell).
	SwitchCommand string `yaml:"switch_command"`
}

// Provider is a single configured git-hosting provider.
type Provider struct {
	// Name is the unique key for this provider entry.
	Name string `yaml:"name"`
	// Type is the provider kind (github, codeberg, gitlab).
	Type ProviderType `yaml:"type"`
	// Short is the path prefix / short code (e.g. gh, cb, gl).
	Short string `yaml:"short"`
	// Username is the authenticated user's own login on this provider (required).
	// It identifies the self account: an empty owners list resolves to it, and
	// provider clients route it to the authenticated-user endpoint. There is no
	// sentinel for "your own account"; the self account is an ordinary owner
	// named by Username.
	Username string `yaml:"username"`
	// APIURL overrides the REST API base URL (e.g. for GHE / self-hosted).
	APIURL string `yaml:"api_url"`
	// WebURL is the base URL used for "open in browser".
	WebURL string `yaml:"web_url"`
	// CloneProtocol is ssh or https; defaults to ssh.
	CloneProtocol string `yaml:"clone_protocol"`
	// ClonePatternTpl optionally overrides the global clone-path template.
	ClonePatternTpl string `yaml:"clone_pattern_tpl"`
	// Auth describes how to obtain a credential for this provider.
	Auth Auth `yaml:"auth"`
	// Owners is an optional allow-list of owners/orgs; empty = all accessible.
	Owners []string `yaml:"owners"`
	// AllOwners, when true, discovers every organization the authenticated user
	// belongs to (plus the user's own account, Username) and unions them with
	// Owners.
	AllOwners bool `yaml:"all_owners"`
	// ExcludeOwners lists owners to ignore (case-insensitive); may include the
	// user's own account by its Username.
	ExcludeOwners []string `yaml:"exclude_owners"`
	// IncludeArchived controls whether archived repos are listed; default false.
	IncludeArchived *bool `yaml:"include_archived"`
	// IncludeForks controls whether fork repos are listed; default true.
	IncludeForks *bool `yaml:"include_forks"`
}

// ResolveOwners computes the effective owner set for a provider given the
// organizations discovered from the provider (may be nil). It is pure. The self
// account participates as an ordinary owner named by p.Username.
//
// When AllOwners is false: the explicit Owners list is used, and an empty list
// resolves to the user's own account (p.Username), with exclude_owners honored.
//
// When AllOwners is true the result is the union of:
//   - p.Username (the user's own account),
//   - every discovered organization,
//   - every explicit Owners entry,
//
// minus ExcludeOwners. Matching for exclusion and de-duplication is
// case-insensitive; the user's own account is excluded when exclude_owners
// contains the Username. The result is de-duplicated, with the Username first
// and the remaining owners sorted.
func ResolveOwners(p *Provider, discovered []string) []string {
	excluded := make(map[string]struct{}, len(p.ExcludeOwners))
	for _, e := range p.ExcludeOwners {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		excluded[strings.ToLower(e)] = struct{}{}
	}
	isExcluded := func(o string) bool {
		_, drop := excluded[strings.ToLower(o)]
		return drop
	}

	if !p.AllOwners {
		// Explicit owners; an empty list means the user's own account.
		src := p.Owners
		if len(src) == 0 {
			src = []string{p.Username}
		}
		out := make([]string, 0, len(src))
		seen := make(map[string]struct{}, len(src))
		for _, o := range src {
			if o == "" || isExcluded(o) {
				continue
			}
			key := strings.ToLower(o)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, o)
		}
		return out
	}

	// AllOwners: union of self, discovered and explicit owners minus excludes.
	includeSelf := p.Username != "" && !isExcluded(p.Username)
	out := make([]string, 0, len(discovered)+len(p.Owners))
	seen := make(map[string]struct{})
	if includeSelf {
		seen[strings.ToLower(p.Username)] = struct{}{}
	}
	for _, group := range [][]string{discovered, p.Owners} {
		for _, o := range group {
			if o == "" || isExcluded(o) {
				continue
			}
			key := strings.ToLower(o)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, o)
		}
	}

	sort.Strings(out)
	if includeSelf {
		out = append([]string{p.Username}, out...)
	}
	return out
}

// Auth describes credential resolution for a provider.
type Auth struct {
	// CLI is the name of a CLI whose token to reuse when present (e.g. gh, glab).
	CLI string `yaml:"cli"`
	// Env is the environment variable holding a PAT fallback.
	Env string `yaml:"env"`
}

// DefaultPath returns the resolved config path using $XDG_CONFIG_HOME when set,
// otherwise ~/.config/huphop/config.yaml.
func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "huphop", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "huphop", "config.yaml"), nil
}

// Load reads, defaults and validates the configuration from the default path.
func Load() (*Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads, defaults and validates the configuration from an explicit path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found at %s", path)
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Parse decodes YAML with strict decoding and applies defaults. It does not
// validate; callers should call Validate.
func Parse(data []byte) (*Config, error) {
	cfg := &Config{}
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if err := cfg.applyDefaults(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// applyDefaults fills omitted values with documented defaults and expands
// the tilde in base_dir.
func (c *Config) applyDefaults() error {
	if c.BaseDir == "" {
		c.BaseDir = "~"
	}
	expanded, err := expandHome(c.BaseDir)
	if err != nil {
		return err
	}
	c.BaseDir = expanded

	if c.ClonePatternTpl == "" {
		c.ClonePatternTpl = DefaultClonePatternTpl
	}

	if c.DefaultMode == "" {
		c.DefaultMode = ModeManagement
	}
	if len(c.Modes) == 0 {
		// Built-in management mode reproducing the standard layout.
		c.Modes = map[string]ModeConfig{
			ModeManagement: {
				Header: []string{"breadcrumb"},
				Footer: []string{"filter", "facet_status", "status_message", "position_indicator", "action_menu"},
			},
		}
	}

	for i := range c.Providers {
		p := &c.Providers[i]
		if p.CloneProtocol == "" {
			p.CloneProtocol = ProtocolSSH
		}
		if p.IncludeArchived == nil {
			b := false
			p.IncludeArchived = &b
		}
		if p.IncludeForks == nil {
			b := true
			p.IncludeForks = &b
		}
	}
	return nil
}

// expandHome expands a leading ~ or ~/ to the user's home directory and returns
// an absolute path.
func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[len("~/"):])
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %q: %w", path, err)
	}
	return abs, nil
}

// Validate checks the configuration and returns the first actionable error.
func (c *Config) Validate() error {
	if len(c.Providers) == 0 {
		return errors.New("at least one provider is required")
	}

	seen := make(map[string]struct{}, len(c.Providers))
	for i := range c.Providers {
		p := &c.Providers[i]
		if p.Name == "" {
			return fmt.Errorf("provider #%d: missing required field \"name\"", i+1)
		}
		if _, dup := seen[p.Name]; dup {
			return fmt.Errorf("duplicate provider name %q", p.Name)
		}
		seen[p.Name] = struct{}{}

		if p.Type == "" {
			return fmt.Errorf("provider %q: missing required field \"type\"", p.Name)
		}
		if !p.Type.valid() {
			return fmt.Errorf("provider %q: unknown type %q (want github, codeberg or gitlab)", p.Name, p.Type)
		}
		if p.Short == "" {
			return fmt.Errorf("provider %q: missing required field \"short\"", p.Name)
		}
		if p.Username == "" {
			return fmt.Errorf("provider %q: missing required field \"username\"", p.Name)
		}
	}

	if c.DefaultMode != "" && len(c.Modes) > 0 {
		if _, ok := c.Modes[c.DefaultMode]; !ok {
			return fmt.Errorf("default_mode %q is not defined in modes", c.DefaultMode)
		}
	}
	for name, mode := range c.Modes {
		seen := make(map[string]struct{})
		for _, slot := range [][]string{mode.Header, mode.Footer} {
			for _, el := range slot {
				if !KnownElements[el] {
					return fmt.Errorf("mode %q: unknown element %q", name, el)
				}
				if _, dup := seen[el]; dup {
					return fmt.Errorf("mode %q: element %q appears more than once", name, el)
				}
				seen[el] = struct{}{}
			}
		}
		if name == ModeMultiplex && mode.SwitchCommand == "" {
			return fmt.Errorf("mode %q: missing required field \"switch_command\"", name)
		}
	}
	return nil
}

func (t ProviderType) valid() bool {
	switch t {
	case ProviderGitHub, ProviderCodeberg, ProviderGitLab:
		return true
	default:
		return false
	}
}
