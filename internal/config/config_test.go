package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validYAML = `
base_dir: ~/
clone_pattern_tpl: "{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}"
providers:
  - name: github-personal
    type: github
    short: gh
    api_url: https://api.github.com
    web_url: https://github.com
    clone_protocol: ssh
    auth:
      cli: gh
      env: HUPHOP_GITHUB_TOKEN
    owners:
      - mipmip
      - TechNative-B-V
    include_archived: false
    include_forks: true
  - name: codeberg
    type: codeberg
    short: cb
    api_url: https://codeberg.org
    web_url: https://codeberg.org
    auth:
      env: HUPHOP_CODEBERG_TOKEN
`

func TestDefaultPath(t *testing.T) {
	t.Run("XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/xdg/conf")
		got, err := DefaultPath()
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join("/xdg/conf", "huphop", "config.yaml")
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("falls back to home", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		home := t.TempDir()
		t.Setenv("HOME", home)
		got, err := DefaultPath()
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(home, ".config", "huphop", "config.yaml")
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestLoadFrom(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name    string
		content string
		write   bool
		wantErr string // substring; "" = no error
	}{
		{
			name:    "valid config",
			content: validYAML,
			write:   true,
		},
		{
			name:    "missing file",
			write:   false,
			wantErr: "not found",
		},
		{
			name:    "malformed yaml",
			content: "base_dir: [unterminated",
			write:   true,
			wantErr: "invalid config",
		},
		{
			name: "unknown key",
			content: `
providers:
  - name: gh
    type: github
    short: gh
    bogus_field: nope
`,
			write:   true,
			wantErr: "invalid config",
		},
		{
			name: "validation failure surfaces",
			content: `
providers: []
`,
			write:   true,
			wantErr: "at least one provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if tt.write {
				if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := LoadFrom(path)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg == nil {
					t.Fatal("expected config, got nil")
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadUsesDefaultPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	dir := filepath.Join(home, ".config", "huphop")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(validYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Providers) != 2 {
		t.Fatalf("got %d providers, want 2", len(cfg.Providers))
	}
}

func TestParseDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := Parse([]byte(`
providers:
  - name: gh
    type: github
    short: gh
`))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.BaseDir != home {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, home)
	}
	if cfg.ClonePatternTpl != DefaultClonePatternTpl {
		t.Errorf("ClonePatternTpl = %q, want default", cfg.ClonePatternTpl)
	}
	p := cfg.Providers[0]
	if p.CloneProtocol != ProtocolSSH {
		t.Errorf("CloneProtocol = %q, want ssh", p.CloneProtocol)
	}
	if p.IncludeArchived == nil || *p.IncludeArchived != false {
		t.Errorf("IncludeArchived = %v, want false", p.IncludeArchived)
	}
	if p.IncludeForks == nil || *p.IncludeForks != true {
		t.Errorf("IncludeForks = %v, want true", p.IncludeForks)
	}
}

func TestBaseDirTildeExpansion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name    string
		baseDir string
		want    string
	}{
		{"bare tilde", "~", home},
		{"tilde slash", "~/", home},
		{"tilde subdir", "~/code", filepath.Join(home, "code")},
		{"absolute untouched", "/opt/repos", "/opt/repos"},
		{"empty defaults to home", "", home},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yml := "providers:\n  - {name: gh, type: github, short: gh}\n"
			if tt.baseDir != "" {
				yml = "base_dir: " + tt.baseDir + "\n" + yml
			}
			cfg, err := Parse([]byte(yml))
			if err != nil {
				t.Fatal(err)
			}
			if cfg.BaseDir != tt.want {
				t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	newBool := func(b bool) *bool { return &b }
	_ = newBool

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid",
			cfg: Config{Providers: []Provider{
				{Name: "gh", Type: ProviderGitHub, Short: "gh"},
			}},
		},
		{
			name:    "no providers",
			cfg:     Config{},
			wantErr: "at least one provider",
		},
		{
			name: "duplicate names",
			cfg: Config{Providers: []Provider{
				{Name: "dup", Type: ProviderGitHub, Short: "gh"},
				{Name: "dup", Type: ProviderCodeberg, Short: "cb"},
			}},
			wantErr: "duplicate provider name",
		},
		{
			name: "unknown type",
			cfg: Config{Providers: []Provider{
				{Name: "bad", Type: "bitbucket", Short: "bb"},
			}},
			wantErr: "unknown type",
		},
		{
			name: "missing name",
			cfg: Config{Providers: []Provider{
				{Type: ProviderGitHub, Short: "gh"},
			}},
			wantErr: `missing required field "name"`,
		},
		{
			name: "missing type",
			cfg: Config{Providers: []Provider{
				{Name: "gh", Short: "gh"},
			}},
			wantErr: `missing required field "type"`,
		},
		{
			name: "missing short",
			cfg: Config{Providers: []Provider{
				{Name: "gh", Type: ProviderGitHub},
			}},
			wantErr: `missing required field "short"`,
		},
		{
			name: "all three types valid",
			cfg: Config{Providers: []Provider{
				{Name: "a", Type: ProviderGitHub, Short: "gh"},
				{Name: "b", Type: ProviderCodeberg, Short: "cb"},
				{Name: "c", Type: ProviderGitLab, Short: "gl"},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidConfigContents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Providers[0].Auth.CLI != "gh" {
		t.Errorf("auth.cli = %q, want gh", cfg.Providers[0].Auth.CLI)
	}
	if len(cfg.Providers[0].Owners) != 2 {
		t.Errorf("owners = %v, want 2", cfg.Providers[0].Owners)
	}
}

func TestResolveOwnersDefaultOff(t *testing.T) {
	// AllOwners false: explicit owners returned as-is (empty means own repos).
	p := &Provider{Owners: []string{"acme", "beta"}}
	got := ResolveOwners(p, []string{"discovered-org"})
	if len(got) != 2 || got[0] != "acme" || got[1] != "beta" {
		t.Fatalf("default-off owners = %v, want [acme beta] (discovery ignored)", got)
	}

	// Empty explicit owners means own repos: a single-element {SelfOwner} set is
	// not synthesized off; the empty list is preserved (clients treat it as own).
	empty := &Provider{}
	if got := ResolveOwners(empty, []string{"x"}); len(got) != 0 {
		t.Fatalf("default-off empty owners = %v, want empty", got)
	}
}

func TestResolveOwnersDefaultOffHonorsExclude(t *testing.T) {
	p := &Provider{Owners: []string{"acme", "beta"}, ExcludeOwners: []string{"BETA"}}
	got := ResolveOwners(p, nil)
	if len(got) != 1 || got[0] != "acme" {
		t.Fatalf("default-off exclude = %v, want [acme]", got)
	}
}

func TestResolveOwnersUnion(t *testing.T) {
	p := &Provider{
		AllOwners: true,
		Owners:    []string{"extra"},
	}
	got := ResolveOwners(p, []string{"orgb", "orga"})
	// Self first (empty), then sorted orgs and explicit owners.
	want := []string{SelfOwner, "extra", "orga", "orgb"}
	if len(got) != len(want) {
		t.Fatalf("union = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("union = %v, want %v", got, want)
		}
	}
}

func TestResolveOwnersUnionDedupCaseInsensitive(t *testing.T) {
	p := &Provider{
		AllOwners: true,
		Owners:    []string{"Acme", "acme"},
	}
	got := ResolveOwners(p, []string{"ACME", "beta"})
	// SelfOwner + acme (first-seen casing) + beta.
	want := []string{SelfOwner, "ACME", "beta"}
	if len(got) != len(want) {
		t.Fatalf("dedup = %v, want %v", got, want)
	}
	// The self sentinel sorts first; the two named owners are sorted, first-wins
	// on casing: discovered "ACME" is seen before explicit "Acme".
	if got[0] != SelfOwner || got[1] != "ACME" || got[2] != "beta" {
		t.Fatalf("dedup = %v, want %v", got, want)
	}
}

func TestResolveOwnersExcludeIncludingSelf(t *testing.T) {
	p := &Provider{
		AllOwners:     true,
		Owners:        []string{"keep"},
		ExcludeOwners: []string{"Noisy", "self"},
	}
	got := ResolveOwners(p, []string{"noisy", "keep", "other"})
	// self excluded via "self" token; "noisy" excluded case-insensitively.
	want := []string{"keep", "other"}
	if len(got) != len(want) {
		t.Fatalf("exclude = %v, want %v", got, want)
	}
	for _, o := range got {
		if o == SelfOwner {
			t.Fatalf("self should have been excluded: %v", got)
		}
	}
	if got[0] != "keep" || got[1] != "other" {
		t.Fatalf("exclude = %v, want %v", got, want)
	}
}
