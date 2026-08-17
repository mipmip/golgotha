package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// Config is the top-level skull2 configuration parsed from config.yaml.
type Config struct {
	// BaseDir is the root under which all trees are created. Tilde-expanded to
	// an absolute path after loading.
	BaseDir string `yaml:"base_dir"`
	// ClonePatternTpl is the global default clone-path template.
	ClonePatternTpl string `yaml:"clone_pattern_tpl"`
	// Providers is the configured list of git providers.
	Providers []Provider `yaml:"providers"`
}

// Provider is a single configured git-hosting provider.
type Provider struct {
	// Name is the unique key for this provider entry.
	Name string `yaml:"name"`
	// Type is the provider kind (github, codeberg, gitlab).
	Type ProviderType `yaml:"type"`
	// Short is the path prefix / short code (e.g. gh, cb, gl).
	Short string `yaml:"short"`
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
	// IncludeArchived controls whether archived repos are listed; default false.
	IncludeArchived *bool `yaml:"include_archived"`
	// IncludeForks controls whether fork repos are listed; default true.
	IncludeForks *bool `yaml:"include_forks"`
}

// Auth describes credential resolution for a provider.
type Auth struct {
	// CLI is the name of a CLI whose token to reuse when present (e.g. gh, glab).
	CLI string `yaml:"cli"`
	// Env is the environment variable holding a PAT fallback.
	Env string `yaml:"env"`
}

// DefaultPath returns the resolved config path using $XDG_CONFIG_HOME when set,
// otherwise ~/.config/skull2/config.yaml.
func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "skull2", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "skull2", "config.yaml"), nil
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
