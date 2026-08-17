package clonepath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mipmip/skull2/internal/config"
)

func TestRenderDefaultTemplate(t *testing.T) {
	p := &config.Provider{Name: "github-personal", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData("/home/pim", p, "github.com", "TechNative-B-V", "foo")
	got, err := Render(config.DefaultClonePatternTpl, data)
	if err != nil {
		t.Fatal(err)
	}
	want := "/home/pim/gh.technative-b-v/foo"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderAllFields(t *testing.T) {
	p := &config.Provider{Name: "github-personal", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData("/base", p, "github.com", "TechNative-B-V", "Foo")

	tpl := "{{.BaseDir}}/{{.Provider}}/{{.Type}}/{{.Short}}/{{.Host}}/{{.Owner}}/{{.OwnerLower}}/{{.Repo}}/{{.RepoLower}}"
	got, err := Render(tpl, data)
	if err != nil {
		t.Fatal(err)
	}
	want := "/base/github-personal/github/gh/github.com/TechNative-B-V/technative-b-v/Foo/foo"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNewDataDerivesLowercase(t *testing.T) {
	p := &config.Provider{Name: "x", Type: config.ProviderGitLab, Short: "gl"}
	d := NewData("/b", p, "h", "MixedCase", "RepoName")
	if d.OwnerLower != "mixedcase" {
		t.Errorf("OwnerLower = %q", d.OwnerLower)
	}
	if d.RepoLower != "reponame" {
		t.Errorf("RepoLower = %q", d.RepoLower)
	}
	if d.Type != "gitlab" {
		t.Errorf("Type = %q", d.Type)
	}
}

func TestTemplateFor(t *testing.T) {
	cfg := &config.Config{ClonePatternTpl: "global"}

	t.Run("global when no override", func(t *testing.T) {
		p := &config.Provider{}
		if got := TemplateFor(cfg, p); got != "global" {
			t.Fatalf("got %q want global", got)
		}
	})
	t.Run("override wins", func(t *testing.T) {
		p := &config.Provider{ClonePatternTpl: "override"}
		if got := TemplateFor(cfg, p); got != "override" {
			t.Fatalf("got %q want override", got)
		}
	})
}

func TestRenderForOverride(t *testing.T) {
	cfg := &config.Config{
		BaseDir:         "/base",
		ClonePatternTpl: config.DefaultClonePatternTpl,
	}
	p := &config.Provider{
		Name:            "gl",
		Type:            config.ProviderGitLab,
		Short:           "gl",
		ClonePatternTpl: "{{.BaseDir}}/custom/{{.RepoLower}}",
	}
	got, err := RenderFor(cfg, p, "gitlab.com", "Group", "Repo")
	if err != nil {
		t.Fatal(err)
	}
	want := "/base/custom/repo"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderTraversalRejected(t *testing.T) {
	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}

	tests := []struct {
		name string
		tpl  string
	}{
		{"parent escape", "{{.BaseDir}}/../evil/{{.Repo}}"},
		{"double parent", "{{.BaseDir}}/../../etc/{{.Repo}}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewData("/home/pim/base", p, "github.com", "owner", "repo")
			_, err := Render(tt.tpl, data)
			if err == nil {
				t.Fatal("expected traversal error, got nil")
			}
			if !strings.Contains(err.Error(), "escapes base_dir") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRenderMissingKeyErrors(t *testing.T) {
	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData("/base", p, "h", "o", "r")
	_, err := Render("{{.BaseDir}}/{{.DoesNotExist}}", data)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestRenderBadTemplate(t *testing.T) {
	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData("/base", p, "h", "o", "r")
	_, err := Render("{{.BaseDir", data)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderTildeExpansion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData(home, p, "github.com", "owner", "repo")
	got, err := Render("~/{{.Short}}.{{.OwnerLower}}/{{.Repo}}", data)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "gh.owner", "repo")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestRenderRelativePathResolvesToAbs(t *testing.T) {
	// A rendered relative path is made absolute against cwd; set base_dir to cwd
	// so it stays within base.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData(cwd, p, "github.com", "owner", "repo")
	got, err := Render("{{.Short}}.{{.OwnerLower}}/{{.Repo}}", data)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(cwd, "gh.owner", "repo")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("result %q is not absolute", got)
	}
}

func TestRenderBaseEqualsTarget(t *testing.T) {
	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData("/home/pim", p, "github.com", "owner", "repo")
	got, err := Render("{{.BaseDir}}", data)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/pim" {
		t.Fatalf("got %q want /home/pim", got)
	}
}

func TestRenderResultUnderBase(t *testing.T) {
	p := &config.Provider{Name: "x", Type: config.ProviderGitHub, Short: "gh"}
	data := NewData("/home/pim", p, "github.com", "owner", "repo")
	got, err := Render(config.DefaultClonePatternTpl, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "/home/pim/") {
		t.Fatalf("result %q not under base", got)
	}
}
