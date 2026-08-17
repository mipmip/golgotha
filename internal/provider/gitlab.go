package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mipmip/golgotha/internal/config"
	"github.com/mipmip/golgotha/internal/fetch"
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
	Visibility     string    `json:"visibility"`
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
// mapped to the Repo model and filtered. It is a thin wrapper over the
// event-emitting FetchOwner.
func (g *GitLab) ListRepos(ctx context.Context, owners []string) ([]Repo, error) {
	return listReposOverFetch(ctx, owners, &g.cfg, g.FetchOwner)
}

// FetchOwner fetches a single group's projects (including subgroups) page by
// page, emitting progress events through emit. An empty owner (config.SelfOwner)
// fetches the authenticated user's membership projects. The total page count is
// derived from the X-Total-Pages header when present.
func (g *GitLab) FetchOwner(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, err
	}

	var path string
	if owner == config.SelfOwner {
		path = "/projects?membership=true"
	} else {
		path = fmt.Sprintf("/groups/%s/projects?include_subgroups=true", url.PathEscape(owner))
	}
	sep := "&"
	if !strings.Contains(path, "?") {
		sep = "?"
	}
	base := fmt.Sprintf("%s%s%sper_page=%d", g.base, path, sep, gitlabPageLimit)

	fetchPage := func(ctx context.Context, page int) (fetch.Page[Repo], error) {
		u := fmt.Sprintf("%s&page=%d", base, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		req.Header.Set("PRIVATE-TOKEN", token)
		req.Header.Set("Accept", "application/json")

		resp, err := g.client.Do(req)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return fetch.Page[Repo]{}, fmt.Errorf("gitlab: %s: %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
		}

		var projects []glProject
		if err := json.Unmarshal(body, &projects); err != nil {
			return fetch.Page[Repo]{}, fmt.Errorf("gitlab: decoding response: %w", err)
		}
		items := make([]Repo, 0, len(projects))
		for _, p := range projects {
			ns := p.Namespace.FullPath
			if ns == "" {
				ns = p.Namespace.Path
			}
			items = append(items, Repo{
				Owner:         ns,
				Name:          p.Path,
				Description:   p.Description,
				SSHURL:        p.SSHURLToRepo,
				HTTPSURL:      p.HTTPURLToRepo,
				WebURL:        p.WebURL,
				DefaultBranch: p.DefaultBranch,
				Archived:      p.Archived,
				Fork:          p.ForkedFromProject != nil,
				Visibility:    NormalizeVisibility(p.Visibility),
				UpdatedAt:     p.LastActivityAt,
			})
		}
		out := fetch.Page[Repo]{Items: items}
		if page == 1 {
			if n, err := strconv.Atoi(resp.Header.Get("X-Total-Pages")); err == nil && n > 0 {
				out.TotalPages = n
			}
		}
		return out, nil
	}

	repos, err := fetch.Pages(ctx, emit, g.cfg.Name, owner, gitlabPageLimit, fetchPage, repoKey)
	if err != nil {
		return nil, err
	}
	return FilterRepos(&g.cfg, repos), nil
}

// glDetail mirrors the subset of the GitLab project JSON consumed for tier-2
// details.
type glDetail struct {
	StarCount     int      `json:"star_count"`
	Topics        []string `json:"topics"`
	DefaultBranch string   `json:"default_branch"`
}

// get performs an authenticated GET against the GitLab API, returning the body
// and status code.
func (g *GitLab) get(ctx context.Context, path string) ([]byte, int, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	body, err := readAllAndClose(resp)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// projectID returns the URL-encoded "owner/name" project identifier the GitLab
// API accepts in place of a numeric ID.
func projectID(owner, name string) string {
	return url.PathEscape(owner + "/" + name)
}

// RepoDetails fetches star_count and topics via GET /projects/{id} and the
// primary language via GET /projects/{id}/languages (the highest-percentage
// entry). A languages fetch failure is not fatal.
func (g *GitLab) RepoDetails(ctx context.Context, owner, name string) (Details, error) {
	id := projectID(owner, name)
	path := "/projects/" + id
	body, status, err := g.get(ctx, path)
	if err != nil {
		return Details{}, err
	}
	if status != http.StatusOK {
		return Details{}, fmt.Errorf("gitlab: %s: %s: %s", path, http.StatusText(status), strings.TrimSpace(string(body)))
	}
	var d glDetail
	if err := json.Unmarshal(body, &d); err != nil {
		return Details{}, fmt.Errorf("gitlab: decoding details: %w", err)
	}
	det := Details{Stars: d.StarCount, Topics: d.Topics}

	langPath := path + "/languages"
	if lb, lstatus, lerr := g.get(ctx, langPath); lerr == nil && lstatus == http.StatusOK {
		det.Language = topLanguage(lb)
	}
	return det, nil
}

// topLanguage decodes a GitLab /languages response (a JSON object of
// language->percentage) and returns the language with the highest percentage, or
// "" when the map is empty or unparseable.
func topLanguage(body []byte) string {
	var langs map[string]float64
	if err := json.Unmarshal(body, &langs); err != nil {
		return ""
	}
	best := ""
	var bestPct float64
	for lang, pct := range langs {
		if pct > bestPct || (pct == bestPct && (best == "" || lang < best)) {
			best = lang
			bestPct = pct
		}
	}
	return best
}

// Readme fetches the raw README.md via the files API: GET
// /projects/{id}/repository/files/README.md/raw?ref={branch}. The default branch
// is discovered from the project details. A not-found yields an empty string and
// nil error.
func (g *GitLab) Readme(ctx context.Context, owner, name string) (string, error) {
	id := projectID(owner, name)

	// Discover the default branch (README raw fetch requires a ref).
	ref := "HEAD"
	if body, status, err := g.get(ctx, "/projects/"+id); err == nil && status == http.StatusOK {
		var d glDetail
		if json.Unmarshal(body, &d) == nil && d.DefaultBranch != "" {
			ref = d.DefaultBranch
		}
	}

	path := fmt.Sprintf("/projects/%s/repository/files/%s/raw?ref=%s",
		id, url.PathEscape("README.md"), url.QueryEscape(ref))
	body, status, err := g.get(ctx, path)
	if err != nil {
		return "", err
	}
	if status == http.StatusNotFound {
		return "", nil
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("gitlab: %s: %s: %s", path, http.StatusText(status), strings.TrimSpace(string(body)))
	}
	return string(body), nil
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
