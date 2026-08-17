package provider

import (
	"io"
	"net/http"
	"os"

	"github.com/mipmip/skull2/internal/config"
)

// osLookupEnv is the production EnvLookup; it mirrors os.LookupEnv.
func osLookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

// readAllAndClose reads the full response body and closes it, always closing the
// body even when reading fails.
func readAllAndClose(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// NewDefaultRegistry returns a Registry with the built-in GitHub, Codeberg and
// GitLab clients registered. Constructors use http.DefaultClient and resolve
// auth via the real CLIs and environment.
func NewDefaultRegistry() *Registry {
	reg := NewRegistry()
	reg.Register(config.ProviderGitHub, func(p *config.Provider) (Provider, error) {
		return NewGitHub(*p, nil), nil
	})
	reg.Register(config.ProviderCodeberg, func(p *config.Provider) (Provider, error) {
		return NewCodeberg(*p, nil), nil
	})
	reg.Register(config.ProviderGitLab, func(p *config.Provider) (Provider, error) {
		return NewGitLab(*p, nil), nil
	})
	return reg
}
