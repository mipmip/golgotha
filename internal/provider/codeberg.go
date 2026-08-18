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

	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
)

// defaultCodebergAPI is the public Codeberg base URL. The Gitea/Forgejo REST
// API lives under /api/v1.
const defaultCodebergAPI = "https://codeberg.org"

// codebergPageLimit is the per-page size requested from the Gitea API.
const codebergPageLimit = 50

// Codeberg lists repositories from a Forgejo/Gitea instance via its REST API.
type Codeberg struct {
	cfg    config.Provider
	client *http.Client
	base   string
	getter CLITokenGetter
	env    EnvLookup
}

// giteaRepo mirrors the subset of the Gitea/Forgejo repository JSON we consume.
type giteaRepo struct {
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

// NewCodeberg builds a Codeberg (Forgejo/Gitea) client. httpClient is injectable
// for tests; a nil client uses http.DefaultClient.
func NewCodeberg(cfg config.Provider, httpClient *http.Client) *Codeberg {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	base := strings.TrimRight(cfg.APIURL, "/")
	if base == "" {
		base = defaultCodebergAPI
	}
	return &Codeberg{
		cfg:    cfg,
		client: httpClient,
		base:   base,
		getter: ExecCLITokenGetter{},
		env:    osLookupEnv,
	}
}

// ListRepos returns repositories for the authenticated user (when no owners are
// configured) plus each configured owner/org, mapped and filtered. It is a thin
// wrapper over the event-emitting FetchOwner.
func (c *Codeberg) ListRepos(ctx context.Context, owners []string) ([]Repo, error) {
	return listReposOverFetch(ctx, owners, &c.cfg, c.FetchOwner)
}

// FetchOwner fetches a single owner's repositories page by page, emitting
// progress events through emit. An empty owner (config.SelfOwner) fetches the
// authenticated user's own repositories. The total page count is derived from
// the X-Total-Count header when present; otherwise pagination falls back to
// sequential (stopping on a short page).
func (c *Codeberg) FetchOwner(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error) {
	token, err := ResolveToken(&c.cfg, c.getter, c.env)
	if err != nil {
		return nil, err
	}

	var path string
	if owner == config.SelfOwner {
		path = "/api/v1/user/repos"
	} else {
		path = fmt.Sprintf("/api/v1/orgs/%s/repos", url.PathEscape(owner))
	}

	fetchPage := func(ctx context.Context, page int) (fetch.Page[Repo], error) {
		u := fmt.Sprintf("%s%s?limit=%d&page=%d", c.base, path, codebergPageLimit, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return fetch.Page[Repo]{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return fetch.Page[Repo]{}, fmt.Errorf("codeberg: %s: %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
		}

		var repos []giteaRepo
		if err := json.Unmarshal(body, &repos); err != nil {
			return fetch.Page[Repo]{}, fmt.Errorf("codeberg: decoding response: %w", err)
		}
		items := make([]Repo, 0, len(repos))
		for _, r := range repos {
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
			out.TotalPages = codebergTotalPages(resp.Header.Get("X-Total-Count"))
		}
		return out, nil
	}

	repos, err := fetch.Pages(ctx, emit, c.cfg.Name, owner, codebergPageLimit, fetchPage, repoKey)
	if err != nil {
		return nil, err
	}
	return FilterRepos(&c.cfg, repos), nil
}

// codebergTotalPages computes the total page count from the X-Total-Count header
// value and the per-page limit. It returns 0 when the header is absent or
// unparseable, signaling the sequential fallback.
func codebergTotalPages(totalCount string) int {
	if totalCount == "" {
		return 0
	}
	n, err := strconv.Atoi(totalCount)
	if err != nil || n <= 0 {
		return 0
	}
	pages := n / codebergPageLimit
	if n%codebergPageLimit != 0 {
		pages++
	}
	if pages < 1 {
		pages = 1
	}
	return pages
}

// giteaDetail mirrors the subset of the Gitea repository JSON consumed for
// tier-2 details.
type giteaDetail struct {
	StarsCount int    `json:"stars_count"`
	Language   string `json:"language"`
}

// giteaTopics mirrors the /topics response.
type giteaTopics struct {
	Topics []string `json:"topics"`
}

// get performs an authenticated GET against the Gitea/Forgejo API, returning the
// body and status code.
func (c *Codeberg) get(ctx context.Context, path string) ([]byte, int, error) {
	token, err := ResolveToken(&c.cfg, c.getter, c.env)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	body, err := readAllAndClose(resp)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// RepoDetails fetches stars and primary language via GET
// /api/v1/repos/{owner}/{name} and topics via the /topics endpoint. A failure to
// fetch topics is not fatal (the repo may have none); the base details still
// return.
func (c *Codeberg) RepoDetails(ctx context.Context, owner, name string) (Details, error) {
	repoPath := fmt.Sprintf("/api/v1/repos/%s/%s", url.PathEscape(owner), url.PathEscape(name))
	body, status, err := c.get(ctx, repoPath)
	if err != nil {
		return Details{}, err
	}
	if status != http.StatusOK {
		return Details{}, fmt.Errorf("codeberg: %s: %s: %s", repoPath, http.StatusText(status), strings.TrimSpace(string(body)))
	}
	var d giteaDetail
	if err := json.Unmarshal(body, &d); err != nil {
		return Details{}, fmt.Errorf("codeberg: decoding details: %w", err)
	}
	det := Details{Stars: d.StarsCount, Language: d.Language}

	topicsPath := repoPath + "/topics"
	if tb, tstatus, terr := c.get(ctx, topicsPath); terr == nil && tstatus == http.StatusOK {
		var t giteaTopics
		if json.Unmarshal(tb, &t) == nil {
			det.Topics = t.Topics
		}
	}
	return det, nil
}

// Readme fetches the raw README markdown via the raw endpoint. Gitea has no
// dedicated README endpoint, so it reads the default branch's README.md. A
// not-found (no README) yields an empty string and nil error.
func (c *Codeberg) Readme(ctx context.Context, owner, name string) (string, error) {
	path := fmt.Sprintf("/api/v1/repos/%s/%s/raw/README.md", url.PathEscape(owner), url.PathEscape(name))
	body, status, err := c.get(ctx, path)
	if err != nil {
		return "", err
	}
	if status == http.StatusNotFound {
		return "", nil
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("codeberg: %s: %s: %s", path, http.StatusText(status), strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

// ListOwners discovers the authenticated user's organizations via
// /api/v1/user/orgs, paginating until a short page. It returns org usernames.
func (c *Codeberg) ListOwners(ctx context.Context) ([]string, error) {
	token, err := ResolveToken(&c.cfg, c.getter, c.env)
	if err != nil {
		return nil, err
	}

	type giteaOrg struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	}
	var out []string
	for page := 1; ; page++ {
		u := fmt.Sprintf("%s/api/v1/user/orgs?limit=%d&page=%d", c.base, codebergPageLimit, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := readAllAndClose(resp)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("codeberg: /api/v1/user/orgs: %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}
		var orgs []giteaOrg
		if err := json.Unmarshal(body, &orgs); err != nil {
			return nil, fmt.Errorf("codeberg: decoding orgs: %w", err)
		}
		for _, o := range orgs {
			name := o.Username
			if name == "" {
				name = o.Name
			}
			out = append(out, name)
		}
		if codebergIsLastPage(resp.Header, page, len(orgs)) {
			break
		}
	}
	return out, nil
}

// codebergIsLastPage reports whether the current page is the final one, using
// the X-Total-Count header when available and falling back to a short page.
func codebergIsLastPage(h http.Header, page, got int) bool {
	if got < codebergPageLimit {
		return true
	}
	if total := h.Get("X-Total-Count"); total != "" {
		if n, err := strconv.Atoi(total); err == nil {
			return page*codebergPageLimit >= n
		}
	}
	return false
}
