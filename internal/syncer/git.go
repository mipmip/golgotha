package syncer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Git abstracts the git operations the sync engine needs so tests can inject a
// fake runner while production shells out to the real git binary.
type Git interface {
	// Clone clones url into dir. dir must not already exist (or be empty).
	Clone(ctx context.Context, url, dir string) error
	// Fetch runs `git fetch` in dir.
	Fetch(ctx context.Context, dir string) error
	// FastForward runs `git merge --ff-only origin/<branch>` in dir.
	FastForward(ctx context.Context, dir, branch string) error
	// IsDirty reports whether `git status --porcelain` is non-empty in dir.
	IsDirty(ctx context.Context, dir string) (bool, error)
	// IsRepo reports whether dir contains a git repository (a .git entry).
	IsRepo(dir string) bool
	// CurrentDefaultBranch returns the remote HEAD branch for dir, used as a
	// fallback when the recorded default branch is unknown.
	CurrentDefaultBranch(ctx context.Context, dir string) (string, error)
}

// ExecGit is the real Git implementation shelling to the `git` binary.
type ExecGit struct {
	// Bin is the git executable to invoke; defaults to "git" when empty.
	Bin string
}

// NewExecGit returns an ExecGit using the git binary on PATH.
func NewExecGit() *ExecGit { return &ExecGit{Bin: "git"} }

func (g *ExecGit) bin() string {
	if g.Bin == "" {
		return "git"
	}
	return g.Bin
}

// run executes git with args, optionally inside dir, and returns stdout. On
// failure it returns an error that includes the trimmed stderr for diagnostics.
func (g *ExecGit) run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, g.bin(), args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, msg)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return stdout.String(), nil
}

// Clone clones url into dir, creating parent directories as needed.
func (g *ExecGit) Clone(ctx context.Context, url, dir string) error {
	if parent := filepath.Dir(dir); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("creating parent dir %s: %w", parent, err)
		}
	}
	_, err := g.run(ctx, "", "clone", url, dir)
	return err
}

// gitProgressRe matches a git `--progress` line like
// "Receiving objects:  58% (580/1000), 1.5 MiB", capturing the phase and the
// integer percentage. An optional "remote: " prefix is tolerated.
var gitProgressRe = regexp.MustCompile(`^(?:remote: )?([A-Za-z][A-Za-z ]+):\s+(\d+)%`)

// parseGitCloneProgress extracts a fraction (0..1) and phase from a single git
// progress line. ok is false for lines that carry no percentage.
func parseGitCloneProgress(line string) (frac float64, phase string, ok bool) {
	m := gitProgressRe.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return 0, "", false
	}
	pct, err := strconv.Atoi(m[2])
	if err != nil {
		return 0, "", false
	}
	return float64(pct) / 100, strings.TrimSpace(m[1]), true
}

// scanCRorLF is a bufio.SplitFunc that splits on either '\r' or '\n', so git's
// carriage-return-updated progress lines are each delivered as a token.
func scanCRorLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b == '\r' || b == '\n' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// CloneProgress clones url into dir like Clone, but runs `git clone --progress`
// and parses git's stderr into fraction/phase progress events emitted via emit
// (which may be nil). It satisfies the optional ProgressGit interface.
func (g *ExecGit) CloneProgress(ctx context.Context, url, dir string, emit func(frac float64, phase string)) error {
	if parent := filepath.Dir(dir); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("creating parent dir %s: %w", parent, err)
		}
	}
	cmd := exec.CommandContext(ctx, g.bin(), "clone", "--progress", url, dir)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	sc := bufio.NewScanner(stderr)
	sc.Split(scanCRorLF)
	for sc.Scan() {
		if frac, phase, ok := parseGitCloneProgress(sc.Text()); ok && emit != nil {
			emit(frac, phase)
		}
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("git clone %s: %w", url, err)
	}
	return nil
}

// Fetch runs `git fetch` in dir.
func (g *ExecGit) Fetch(ctx context.Context, dir string) error {
	_, err := g.run(ctx, dir, "fetch")
	return err
}

// FastForward runs `git merge --ff-only origin/<branch>` in dir.
func (g *ExecGit) FastForward(ctx context.Context, dir, branch string) error {
	_, err := g.run(ctx, dir, "merge", "--ff-only", "origin/"+branch)
	return err
}

// IsDirty reports whether the working tree has uncommitted changes.
func (g *ExecGit) IsDirty(ctx context.Context, dir string) (bool, error) {
	out, err := g.run(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// IsRepo reports whether dir contains a .git entry.
func (g *ExecGit) IsRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode().IsRegular()
}

// CurrentDefaultBranch resolves the remote HEAD branch (e.g. "main") for dir.
func (g *ExecGit) CurrentDefaultBranch(ctx context.Context, dir string) (string, error) {
	out, err := g.run(ctx, dir, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(out)
	// ref looks like "origin/main"; strip the remote prefix.
	if i := strings.IndexByte(ref, '/'); i >= 0 {
		ref = ref[i+1:]
	}
	if ref == "" {
		return "", fmt.Errorf("could not determine default branch for %s", dir)
	}
	return ref, nil
}
