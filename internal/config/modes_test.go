package config

import (
	"strings"
	"testing"
)

func TestModesDefaultWhenOmitted(t *testing.T) {
	c := &Config{Providers: []Provider{{Name: "gh", Type: ProviderGitHub, Short: "gh", Username: "me"}}}
	if err := c.applyDefaults(); err != nil {
		t.Fatalf("applyDefaults: %v", err)
	}
	if c.DefaultMode != ModeManagement {
		t.Fatalf("default mode = %q, want %q", c.DefaultMode, ModeManagement)
	}
	mgmt, ok := c.Modes[ModeManagement]
	if !ok {
		t.Fatalf("built-in management mode missing: %+v", c.Modes)
	}
	if len(mgmt.Header) != 1 || mgmt.Header[0] != "breadcrumb" {
		t.Fatalf("management header = %v", mgmt.Header)
	}
	if len(mgmt.Footer) != 5 || mgmt.Footer[4] != "action_menu" {
		t.Fatalf("management footer = %v", mgmt.Footer)
	}
}

func TestModesValidation(t *testing.T) {
	base := func() Config {
		return Config{Providers: []Provider{{Name: "gh", Type: ProviderGitHub, Short: "gh", Username: "me"}}}
	}
	cases := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name: "unknown element",
			mutate: func(c *Config) {
				c.DefaultMode = "management"
				c.Modes = map[string]ModeConfig{"management": {Header: []string{"bogus"}}}
			},
			wantErr: `unknown element "bogus"`,
		},
		{
			name: "duplicate element per mode",
			mutate: func(c *Config) {
				c.DefaultMode = "management"
				c.Modes = map[string]ModeConfig{"management": {
					Header: []string{"breadcrumb"},
					Footer: []string{"breadcrumb"},
				}}
			},
			wantErr: `appears more than once`,
		},
		{
			name: "default_mode not defined",
			mutate: func(c *Config) {
				c.DefaultMode = "nope"
				c.Modes = map[string]ModeConfig{"management": {Header: []string{"breadcrumb"}}}
			},
			wantErr: `default_mode "nope" is not defined`,
		},
		{
			name: "multiplex needs switch_command",
			mutate: func(c *Config) {
				c.DefaultMode = "management"
				c.Modes = map[string]ModeConfig{
					"management": {Header: []string{"breadcrumb"}},
					"multiplex":  {Footer: []string{"switch_hint"}},
				}
			},
			wantErr: `mode "multiplex": missing required field "switch_command"`,
		},
		{
			name: "valid modes pass",
			mutate: func(c *Config) {
				c.DefaultMode = "multiplex"
				c.Modes = map[string]ModeConfig{
					"management": {Header: []string{"breadcrumb"}, Footer: []string{"action_menu"}},
					"multiplex":  {Footer: []string{"switch_hint"}, SwitchCommand: "echo {{.Repo}}"},
				}
			},
			wantErr: "",
		},
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
