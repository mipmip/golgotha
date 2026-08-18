package provider

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mipmip/huphop/internal/config"
)

// EnvLookup looks up an environment variable, returning its value and whether
// it was set. It mirrors os.LookupEnv and is injectable for hermetic tests.
type EnvLookup func(key string) (string, bool)

// CLITokenGetter obtains an auth token from a provider CLI (e.g. gh, glab).
// Token returns the token and true when a CLI token is available, or ("", false)
// with a nil error when the CLI is present but has no usable token. A non-nil
// error indicates the CLI could not be consulted.
type CLITokenGetter interface {
	Token(cli string) (string, bool, error)
}

// ResolveToken resolves a provider's credential in order: configured CLI token,
// then env-var PAT, then an actionable error naming the provider and env var.
//
// getter and env are injectable so unit tests avoid shelling out and the real
// environment; pass ExecCLITokenGetter{} and os.LookupEnv in production.
func ResolveToken(p *config.Provider, getter CLITokenGetter, env EnvLookup) (string, error) {
	if p.Auth.CLI != "" && getter != nil {
		token, ok, err := getter.Token(p.Auth.CLI)
		if err == nil && ok && token != "" {
			return token, nil
		}
	}

	if p.Auth.Env != "" && env != nil {
		if val, ok := env(p.Auth.Env); ok && val != "" {
			return val, nil
		}
	}

	return "", authError(p)
}

func authError(p *config.Provider) error {
	if p.Auth.Env != "" {
		return fmt.Errorf("no credential for provider %q: set the %s environment variable", p.Name, p.Auth.Env)
	}
	if p.Auth.CLI != "" {
		return fmt.Errorf("no credential for provider %q: authenticate the %s CLI or configure an env PAT", p.Name, p.Auth.CLI)
	}
	return fmt.Errorf("no credential for provider %q: configure auth.cli or auth.env", p.Name)
}

// ExecCLITokenGetter obtains tokens by shelling out to the provider CLI.
type ExecCLITokenGetter struct{}

// Token runs `<cli> auth token` (gh/glab convention) and returns the trimmed
// output. A non-zero exit is treated as "no token available", not an error.
func (ExecCLITokenGetter) Token(cli string) (string, bool, error) {
	path, err := exec.LookPath(cli)
	if err != nil {
		return "", false, nil
	}
	out, err := exec.Command(path, "auth", "token").Output()
	if err != nil {
		return "", false, nil
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", false, nil
	}
	return token, true, nil
}
