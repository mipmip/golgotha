package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
)

// defaultGitHubAPI is the public GitHub REST API base URL.
const defaultGitHubAPI = "https://api.github.com"

// githubPageLimit is the per-page size requested from the GitHub API.
const githubPageLimit = 100

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
	Private       bool      `json:"private"`
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
// filtered per include_archived / include_forks. It is a thin wrapper over the
// event-emitting FetchOwner, invoked once per owner with no event consumer.
func (g *GitHub) ListRepos(ctx context.Context, owners []string) ([]Repo, error) {
	return listReposOverFetch(ctx, owners, &g.cfg, g.FetchOwner)
}

// FetchOwner fetches a single owner's repositories page by page, emitting
// progress events through emit. When owner equals the configured Username it
// fetches the authenticated user's own repositories; any other owner is fetched
// as an organization. Results are mapped and filtered per the provider config.
// The cache/caller is responsible for committing only on a complete (non-error,
// non-canceled) result.
func (g *GitHub) FetchOwner(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, err
	}

	var pathBase string
	if owner == g.cfg.Username {
		pathBase = fmt.Sprintf("/user/repos?per_page=%d&affiliation=owner", githubPageLimit)
	} else {
		pathBase = fmt.Sprintf("/orgs/%s/repos?per_page=%d", url.PathEscape(owner), githubPageLimit)
	}

	fetchPage := func(ctx context.Context, page int) (fetch.Page[Repo], error) {
		u := fmt.Sprintf("%s%s&page=%d", g.base, pathBase, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := g.client.Do(req)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return fetch.Page[Repo]{}, fmt.Errorf("github: %s: %s: %s", req.URL.Path, resp.Status, strings.TrimSpace(string(body)))
		}

		var pageRepos []ghRepo
		if err := json.Unmarshal(body, &pageRepos); err != nil {
			return fetch.Page[Repo]{}, fmt.Errorf("github: decoding response: %w", err)
		}
		items := make([]Repo, 0, len(pageRepos))
		for _, r := range pageRepos {
			items = append(items, Repo{
				Owner:         r.Owner.Login,
				Name:          r.Name,
				Description:   r.Description,
				SSHURL:        r.SSHURL,
				HTTPSURL:      r.CloneURL,
				WebURL:        r.HTMLURL,
				DefaultBranch: r.DefaultBranch,
				Archived:      r.Archived,
				Fork:          r.Fork,
				Visibility:    visibilityFromPrivate(r.Private),
				UpdatedAt:     r.UpdatedAt,
			})
		}
		out := fetch.Page[Repo]{Items: items}
		if page == 1 {
			out.TotalPages = lastLinkPage(resp.Header.Get("Link"))
		}
		return out, nil
	}

	repos, err := fetch.Pages(ctx, emit, g.cfg.Name, owner, githubPageLimit, fetchPage, repoKey)
	if err != nil {
		return nil, err
	}
	return FilterRepos(&g.cfg, repos), nil
}

// ListOwners discovers the authenticated user's organizations via /user/orgs,
// following Link-header pagination. It returns the org login names.
func (g *GitHub) ListOwners(ctx context.Context) ([]string, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, err
	}

	type ghOrg struct {
		Login string `json:"login"`
	}
	next := g.base + "/user/orgs?per_page=100"
	var out []string
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
		var page []ghOrg
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("github: decoding orgs: %w", err)
		}
		for _, o := range page {
			out = append(out, o.Login)
		}
		next = nextLink(resp.Header.Get("Link"))
	}
	return out, nil
}

// ghDetail mirrors the subset of the GitHub repository JSON consumed for tier-2
// details. Topics require the mercy preview / topics array on the repo object,
// which the modern API returns by default.
type ghDetail struct {
	StargazersCount int      `json:"stargazers_count"`
	Language        string   `json:"language"`
	Topics          []string `json:"topics"`
}

// ghReadme mirrors the GitHub contents/readme response: base64-encoded content.
type ghReadme struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// get performs an authenticated GET against the GitHub API and returns the body
// and status code. It is shared by the detail/README fetchers.
func (g *GitHub) get(ctx context.Context, path string) ([]byte, int, error) {
	token, err := ResolveToken(&g.cfg, g.getter, g.env)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.base+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

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

// RepoDetails fetches stars, topics and primary language for a GitHub repo via
// GET /repos/{owner}/{name}.
func (g *GitHub) RepoDetails(ctx context.Context, owner, name string) (Details, error) {
	path := fmt.Sprintf("/repos/%s/%s", url.PathEscape(owner), url.PathEscape(name))
	body, status, err := g.get(ctx, path)
	if err != nil {
		return Details{}, err
	}
	if status != http.StatusOK {
		return Details{}, fmt.Errorf("github: %s: %s: %s", path, http.StatusText(status), strings.TrimSpace(string(body)))
	}
	var d ghDetail
	if err := json.Unmarshal(body, &d); err != nil {
		return Details{}, fmt.Errorf("github: decoding details: %w", err)
	}
	return Details{Stars: d.StargazersCount, Topics: d.Topics, Language: d.Language}, nil
}

// Readme fetches the raw README markdown via GET /repos/{owner}/{name}/readme,
// which returns the default README (base64-encoded). A 404 (no README) yields an
// empty string and nil error.
func (g *GitHub) Readme(ctx context.Context, owner, name string) (string, error) {
	path := fmt.Sprintf("/repos/%s/%s/readme", url.PathEscape(owner), url.PathEscape(name))
	body, status, err := g.get(ctx, path)
	if err != nil {
		return "", err
	}
	if status == http.StatusNotFound {
		return "", nil
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("github: %s: %s: %s", path, http.StatusText(status), strings.TrimSpace(string(body)))
	}
	var r ghReadme
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("github: decoding readme: %w", err)
	}
	if strings.EqualFold(r.Encoding, "base64") {
		// GitHub wraps base64 content at 60 cols with newlines.
		clean := strings.ReplaceAll(r.Content, "\n", "")
		decoded, derr := base64.StdEncoding.DecodeString(clean)
		if derr != nil {
			return "", fmt.Errorf("github: decoding readme content: %w", derr)
		}
		return string(decoded), nil
	}
	return r.Content, nil
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

// lastLinkPage extracts the page number from the rel="last" URL of an RFC 5988
// Link header. It returns 0 when there is no rel="last" (single page, or the
// header is absent), which signals the caller to fall back to sequential
// pagination.
func lastLinkPage(header string) int {
	if header == "" {
		return 0
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
		isLast := false
		for _, param := range segs[1:] {
			param = strings.TrimSpace(param)
			if param == `rel="last"` || param == "rel=last" {
				isLast = true
				break
			}
		}
		if !isLast {
			continue
		}
		u, err := url.Parse(rawURL)
		if err != nil {
			return 0
		}
		if n, err := strconv.Atoi(u.Query().Get("page")); err == nil {
			return n
		}
	}
	return 0
}
