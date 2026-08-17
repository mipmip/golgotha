package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mipmip/skull2/internal/config"
)

// defaultGitLabAPI is the public GitLab v4 API base URL.
const defaultGitLabAPI = "https://gitlab.com/api/v4"

// gitlabPageLimit is the per-page size requested from the GitLab API.
const gitlabPageLimit = 100

// GitLab lists projects from a GitLab instance via the v4 REST API.
type GitLab struct {
	cfg    config.Provider
	client *http.Client
	base   string
	getter CLITokenGetter
	env    EnvLookup
}

// glProject mirrors the subset of the GitLab project JSON we consume.
type glProject struct {
	Name           string    `json:"name"`
	Path           string    `json:"path"`
	Description    string    `json:"description"`
	SSHURLToRepo   string    `json:"ssh_url_to_repo"`
	HTTPURLToRepo  string    `json:"http_url_to_repo"`
	WebURL         string    `json:"web_url"`
	DefaultBranch  string    `json:"default_branch"`
	Archived       bool      `json:"archived"`
	LastActivityAt time.Time `json:"last_activity_at"`
	Namespace      struct {
		FullPath string `json:"full_path"`
		Path     string `json:"path"`
	} `json:"namespace"`
	ForkedFromProject *struct {
		ID int `json:"id"`
	} `json:"forked_from_project"`
}

// NewGitLab builds a GitLab client. httpClient is injectable for tests; a nil
// client uses http.DefaultClient.
func NewGitLab(cfg config.Provider, httpClient *http.Client) *GitLab {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	base := strings.TrimRight(cfg.APIURL, "/")
	if base == "" {
		base = defaultGitLabAPI
	}
	return &GitLab{
		cfg:    cfg,
		client: httpClient,
		base:   base,
		getter: ExecCLITokenGetter{},
		env:    osLookupEnv,
	}
}

// ListRepos returns projects for each configured group (including subgroups)
// plus the authenticated user's membership projects when no owners are set,
// mapped to the Repo model and filtered.
func (g *GitLab) ListRepos(ctx context.Context, owners []string) ([]Repo, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, err
	}

	var paths []string
	if len(owners) == 0 {
		paths = append(paths, "/projects?membership=true")
	} else {
		for _, owner := range owners {
			paths = append(paths, fmt.Sprintf("/groups/%s/projects?include_subgroups=true", url.PathEscape(owner)))
		}
	}

	var all []Repo
	seen := make(map[string]struct{})
	for _, p := range paths {
		repos, err := g.listPath(ctx, token, p)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			key := r.Owner + "/" + r.Name
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, r)
		}
	}

	return FilterRepos(&g.cfg, all), nil
}

// ListOwners discovers the groups the authenticated user is a member of via
// /groups?min_access_level=10 (Guest and above), paginating via X-Next-Page. It
// returns each group's full_path so it maps to the owner strings ListRepos uses
// (which fetches /groups/<owner>/projects?include_subgroups=true).
func (g *GitLab) ListOwners(ctx context.Context) ([]string, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, err
	}

	type glGroup struct {
		FullPath string `json:"full_path"`
		Path     string `json:"path"`
	}
	base := fmt.Sprintf("%s/groups?min_access_level=10&all_available=false&per_page=%d", g.base, gitlabPageLimit)

	var out []string
	page := "1"
	for page != "" {
		u := base + "&page=" + url.QueryEscape(page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("PRIVATE-TOKEN", token)
		req.Header.Set("Accept", "application/json")

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gitlab: /groups: %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}
		var groups []glGroup
		if err := json.Unmarshal(body, &groups); err != nil {
			return nil, fmt.Errorf("gitlab: decoding groups: %w", err)
		}
		for _, gr := range groups {
			name := gr.FullPath
			if name == "" {
				name = gr.Path
			}
			out = append(out, name)
		}
		page = resp.Header.Get("X-Next-Page")
	}
	return out, nil
}

// listPath fetches every page for a v4 endpoint path (with an existing query),
// paginating via the X-Next-Page header.
func (g *GitLab) listPath(ctx context.Context, token, path string) ([]Repo, error) {
	sep := "&"
	if !strings.Contains(path, "?") {
		sep = "?"
	}
	base := fmt.Sprintf("%s%s%sper_page=%d", g.base, path, sep, gitlabPageLimit)

	var out []Repo
	page := "1"
	for page != "" {
		u := base + "&page=" + url.QueryEscape(page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("PRIVATE-TOKEN", token)
		req.Header.Set("Accept", "application/json")

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gitlab: %s: %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
		}

		var projects []glProject
		if err := json.Unmarshal(body, &projects); err != nil {
			return nil, fmt.Errorf("gitlab: decoding response: %w", err)
		}
		for _, p := range projects {
			owner := p.Namespace.FullPath
			if owner == "" {
				owner = p.Namespace.Path
			}
			out = append(out, Repo{
				Owner:         owner,
				Name:          p.Path,
				Description:   p.Description,
				SSHURL:        p.SSHURLToRepo,
				HTTPSURL:      p.HTTPURLToRepo,
				WebURL:        p.WebURL,
				DefaultBranch: p.DefaultBranch,
				Archived:      p.Archived,
				Fork:          p.ForkedFromProject != nil,
				UpdatedAt:     p.LastActivityAt,
			})
		}

		page = resp.Header.Get("X-Next-Page")
	}
	return out, nil
}
