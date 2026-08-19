package config

import (
	"strings"
	"testing"
)

func TestCloneVCSForResolution(t *testing.T) {
	p := &Provider{
		Name:     "gh",
		CloneVCS: "git",
		VCSRules: []VCSRule{
			{Match: "mipmip/dotfiles", VCS: "jj"},
			{Match: "mipmip/*", VCS: "jj"},
		},
	}
	c := &Config{CloneVCS: "git"}

	cases := []struct {
		owner string
		want  string
	}{
		{"mipmip/dotfiles", "jj"}, // exact rule
		{"mipmip/other", "jj"},    // glob rule
		{"acme/widget", "git"},    // no rule → provider clone_vcs
	}
	for _, tc := range cases {
		if got := c.CloneVCSFor(p, tc.owner); got != tc.want {
			t.Fatalf("CloneVCSFor(%q) = %q, want %q", tc.owner, got, tc.want)
		}
	}

	// No provider override → falls to global.
	global := &Config{CloneVCS: "jj"}
	if got := global.CloneVCSFor(&Provider{Name: "x"}, "a/b"); got != "jj" {
		t.Fatalf("global fallback = %q, want jj", got)
	}
	// Nothing set anywhere → git.
	if got := (&Config{}).CloneVCSFor(&Provider{Name: "x"}, "a/b"); got != "git" {
		t.Fatalf("default = %q, want git", got)
	}
	// Provider override beats global.
	if got := (&Config{CloneVCS: "jj"}).CloneVCSFor(&Provider{Name: "x", CloneVCS: "git"}, "a/b"); got != "git" {
		t.Fatalf("provider override = %q, want git", got)
	}
}

func TestCloneVCSValidation(t *testing.T) {
	base := func() Config {
		return Config{Providers: []Provider{{Name: "gh", Type: ProviderGitHub, Short: "gh", Username: "me"}}}
	}
	cases := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{"bad global vcs", func(c *Config) { c.CloneVCS = "hg" }, `unknown vcs "hg"`},
		{"bad provider vcs", func(c *Config) { c.Providers[0].CloneVCS = "svn" }, `unknown vcs "svn"`},
		{"rule empty vcs", func(c *Config) {
			c.Providers[0].VCSRules = []VCSRule{{Match: "a/*", VCS: ""}}
		}, `vcs is required`},
		{"rule bad glob", func(c *Config) {
			c.Providers[0].VCSRules = []VCSRule{{Match: "a[", VCS: "jj"}}
		}, `invalid match`},
		{"valid jj rule", func(c *Config) {
			c.CloneVCS = "git"
			c.Providers[0].VCSRules = []VCSRule{{Match: "me/*", VCS: "jj"}}
		}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := base()
			tc.mutate(&c)
			err := c.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}
