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

// defaultGitHubAPI is the public GitHub REST API base URL.
const defaultGitHubAPI = "https://api.github.com"

// GitHub lists repositories from a GitHub (or GitHub Enterprise) instance via
// the REST API.
type GitHub struct {
	cfg    config.Provider
	client *http.Client
	base   string
	getter CLITokenGetter
	env    EnvLookup
}

// ghRepo mirrors the subset of the GitHub repository JSON we consume.
type ghRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	SSHURL        string    `json:"ssh_url"`
	CloneURL      string    `json:"clone_url"`
	HTMLURL       string    `json:"html_url"`
	DefaultBranch string    `json:"default_branch"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewGitHub builds a GitHub client. httpClient and the resolved base URL are
// injectable so tests can point at an httptest.Server. A nil httpClient uses
// http.DefaultClient.
func NewGitHub(cfg config.Provider, httpClient *http.Client) *GitHub {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	base := strings.TrimRight(cfg.APIURL, "/")
	if base == "" {
		base = defaultGitHubAPI
	}
	return &GitHub{
		cfg:    cfg,
		client: httpClient,
		base:   base,
		getter: ExecCLITokenGetter{},
		env:    osLookupEnv,
	}
}

// ListRepos returns repositories for the authenticated user (when no owners are
// configured) plus each configured owner/org, mapped to the Repo model and
// filtered per include_archived / include_forks.
func (g *GitHub) ListRepos(ctx context.Context, owners []string) ([]Repo, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, err
	}

	var paths []string
	if len(owners) == 0 {
		// Authenticated user's own repositories.
		paths = append(paths, "/user/repos?per_page=100&affiliation=owner")
	} else {
		for _, owner := range owners {
			paths = append(paths, fmt.Sprintf("/orgs/%s/repos?per_page=100", url.PathEscape(owner)))
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

// listPath fetches every page starting at path (an absolute API path with query),
// following the Link header for pagination.
func (g *GitHub) listPath(ctx context.Context, token, path string) ([]Repo, error) {
	next := g.base + path
	var out []Repo
	for next != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("github: %s: %s: %s", req.URL.Path, resp.Status, strings.TrimSpace(string(body)))
		}

		var page []ghRepo
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("github: decoding response: %w", err)
		}
		for _, r := range page {
			out = append(out, Repo{
				Owner:         r.Owner.Login,
				Name:          r.Name,
				Description:   r.Description,
				SSHURL:        r.SSHURL,
				HTTPSURL:      r.CloneURL,
				WebURL:        r.HTMLURL,
				DefaultBranch: r.DefaultBranch,
				Archived:      r.Archived,
				Fork:          r.Fork,
				UpdatedAt:     r.UpdatedAt,
			})
		}

		next = nextLink(resp.Header.Get("Link"))
	}
	return out, nil
}

// nextLink extracts the rel="next" URL from an RFC 5988 Link header, or "".
func nextLink(header string) string {
	if header == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		segs := strings.Split(strings.TrimSpace(part), ";")
		if len(segs) < 2 {
			continue
		}
		rawURL := strings.TrimSpace(segs[0])
		if !strings.HasPrefix(rawURL, "<") || !strings.HasSuffix(rawURL, ">") {
			continue
		}
		rawURL = rawURL[1 : len(rawURL)-1]
		for _, param := range segs[1:] {
			param = strings.TrimSpace(param)
			if param == `rel="next"` || param == "rel=next" {
				return rawURL
			}
		}
	}
	return ""
}
