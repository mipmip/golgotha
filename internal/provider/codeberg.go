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

	"github.com/mipmip/skull2/internal/config"
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
// configured) plus each configured owner/org, mapped and filtered.
func (c *Codeberg) ListRepos(ctx context.Context, owners []string) ([]Repo, error) {
	token, err := ResolveToken(&c.cfg, c.getter, c.env)
	if err != nil {
		return nil, err
	}

	var paths []string
	if len(owners) == 0 {
		paths = append(paths, "/api/v1/user/repos")
	} else {
		for _, owner := range owners {
			paths = append(paths, fmt.Sprintf("/api/v1/orgs/%s/repos", url.PathEscape(owner)))
		}
	}

	var all []Repo
	seen := make(map[string]struct{})
	for _, p := range paths {
		repos, err := c.listPath(ctx, token, p)
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

	return FilterRepos(&c.cfg, all), nil
}

// listPath fetches every page for path, paginating via ?page/&limit until a
// short (or empty) page is returned.
func (c *Codeberg) listPath(ctx context.Context, token, path string) ([]Repo, error) {
	var out []Repo
	for page := 1; ; page++ {
		u := fmt.Sprintf("%s%s?limit=%d&page=%d", c.base, path, codebergPageLimit, page)
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
			return nil, fmt.Errorf("codeberg: %s: %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
		}

		var repos []giteaRepo
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("codeberg: decoding response: %w", err)
		}
		for _, r := range repos {
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

		// Stop when this page is the last: prefer X-Total-Count when present,
		// otherwise stop on a short page.
		if last := codebergIsLastPage(resp.Header, page, len(repos)); last {
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
