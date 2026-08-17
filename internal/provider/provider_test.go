package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mipmip/skull2/internal/config"
)

func boolPtr(b bool) *bool { return &b }

// fakeProvider is a hermetic Provider used in tests.
type fakeProvider struct {
	repos []Repo
	err   error
}

func (f *fakeProvider) ListRepos(_ context.Context, _ []string) ([]Repo, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.repos, nil
}

func TestProviderInterfaceListRepos(t *testing.T) {
	want := []Repo{
		{Owner: "o", Name: "a", DefaultBranch: "main", UpdatedAt: time.Now()},
	}
	var p Provider = &fakeProvider{repos: want}
	got, err := p.ListRepos(context.Background(), []string{"o"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("unexpected repos: %+v", got)
	}
}

func TestProviderListReposError(t *testing.T) {
	sentinel := errors.New("boom")
	var p Provider = &fakeProvider{err: sentinel}
	_, err := p.ListRepos(context.Background(), nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want %v", err, sentinel)
	}
}

func TestFilterRepos(t *testing.T) {
	repos := []Repo{
		{Name: "plain"},
		{Name: "arch", Archived: true},
		{Name: "fork", Fork: true},
		{Name: "archfork", Archived: true, Fork: true},
	}

	tests := []struct {
		name            string
		includeArchived *bool
		includeForks    *bool
		want            []string
	}{
		{
			name: "defaults nil: exclude archived, include forks",
			want: []string{"plain", "fork"},
		},
		{
			name:            "include everything",
			includeArchived: boolPtr(true),
			includeForks:    boolPtr(true),
			want:            []string{"plain", "arch", "fork", "archfork"},
		},
		{
			name:            "exclude forks only",
			includeArchived: boolPtr(true),
			includeForks:    boolPtr(false),
			want:            []string{"plain", "arch"},
		},
		{
			name:            "exclude both",
			includeArchived: boolPtr(false),
			includeForks:    boolPtr(false),
			want:            []string{"plain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &config.Provider{IncludeArchived: tt.includeArchived, IncludeForks: tt.includeForks}
			got := FilterRepos(p, repos)
			var names []string
			for _, r := range got {
				names = append(names, r.Name)
			}
			if strings.Join(names, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("got %v want %v", names, tt.want)
			}
		})
	}
}

// stubGetter is an injectable CLITokenGetter for tests.
type stubGetter struct {
	token string
	ok    bool
	err   error
}

func (s stubGetter) Token(_ string) (string, bool, error) {
	return s.token, s.ok, s.err
}

func env(m map[string]string) EnvLookup {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestResolveToken(t *testing.T) {
	tests := []struct {
		name    string
		prov    config.Provider
		getter  CLITokenGetter
		env     map[string]string
		want    string
		wantErr string
	}{
		{
			name:   "cli token used",
			prov:   config.Provider{Name: "gh", Auth: config.Auth{CLI: "gh", Env: "SKULL2_GH"}},
			getter: stubGetter{token: "cli-tok", ok: true},
			env:    map[string]string{"SKULL2_GH": "env-tok"},
			want:   "cli-tok",
		},
		{
			name:   "env fallback when cli has no token",
			prov:   config.Provider{Name: "gh", Auth: config.Auth{CLI: "gh", Env: "SKULL2_GH"}},
			getter: stubGetter{ok: false},
			env:    map[string]string{"SKULL2_GH": "env-tok"},
			want:   "env-tok",
		},
		{
			name:   "env fallback when cli errors",
			prov:   config.Provider{Name: "gh", Auth: config.Auth{CLI: "gh", Env: "SKULL2_GH"}},
			getter: stubGetter{err: errors.New("gh missing")},
			env:    map[string]string{"SKULL2_GH": "env-tok"},
			want:   "env-tok",
		},
		{
			name: "env only provider",
			prov: config.Provider{Name: "cb", Auth: config.Auth{Env: "SKULL2_CB"}},
			env:  map[string]string{"SKULL2_CB": "cb-tok"},
			want: "cb-tok",
		},
		{
			name:    "no credential names env var",
			prov:    config.Provider{Name: "gh", Auth: config.Auth{CLI: "gh", Env: "SKULL2_GH"}},
			getter:  stubGetter{ok: false},
			env:     map[string]string{},
			wantErr: "SKULL2_GH",
		},
		{
			name:    "no credential cli only",
			prov:    config.Provider{Name: "gh", Auth: config.Auth{CLI: "gh"}},
			getter:  stubGetter{ok: false},
			env:     map[string]string{},
			wantErr: "gh CLI",
		},
		{
			name:    "no auth configured",
			prov:    config.Provider{Name: "gh"},
			env:     map[string]string{},
			wantErr: "auth.cli or auth.env",
		},
		{
			name:    "empty env value falls through to error",
			prov:    config.Provider{Name: "gh", Auth: config.Auth{Env: "SKULL2_GH"}},
			env:     map[string]string{"SKULL2_GH": ""},
			wantErr: "SKULL2_GH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveToken(&tt.prov, tt.getter, env(tt.env))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestExecCLITokenGetterMissingCLI(t *testing.T) {
	// A CLI name that will not be found on PATH returns (false, nil).
	tok, ok, err := ExecCLITokenGetter{}.Token("definitely-not-a-real-cli-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok || tok != "" {
		t.Fatalf("expected no token, got %q ok=%v", tok, ok)
	}
}

func TestRegistry(t *testing.T) {
	const fakeType config.ProviderType = "fake"

	reg := NewRegistry()
	built := &fakeProvider{}
	reg.Register(fakeType, func(_ *config.Provider) (Provider, error) {
		return built, nil
	})

	t.Run("known type resolves", func(t *testing.T) {
		p, err := reg.Build(&config.Provider{Type: fakeType})
		if err != nil {
			t.Fatal(err)
		}
		if p != built {
			t.Fatal("built provider mismatch")
		}
	})

	t.Run("unknown type errors", func(t *testing.T) {
		_, err := reg.Build(&config.Provider{Type: "nope"})
		if err == nil {
			t.Fatal("expected error for unknown type")
		}
		if !strings.Contains(err.Error(), "nope") {
			t.Fatalf("error %q does not name the type", err)
		}
	})

	t.Run("constructor error propagates", func(t *testing.T) {
		const errType config.ProviderType = "errtype"
		reg.Register(errType, func(_ *config.Provider) (Provider, error) {
			return nil, errors.New("construct fail")
		})
		_, err := reg.Build(&config.Provider{Type: errType})
		if err == nil || !strings.Contains(err.Error(), "construct fail") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRegistryDuplicatePanics(t *testing.T) {
	const fakeType config.ProviderType = "dup"
	reg := NewRegistry()
	reg.Register(fakeType, func(_ *config.Provider) (Provider, error) { return &fakeProvider{}, nil })

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	reg.Register(fakeType, func(_ *config.Provider) (Provider, error) { return &fakeProvider{}, nil })
}
