package syncer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/creack/pty"
)

// ensureJJ reports an actionable error when the `jj` binary is not on PATH.
func ensureJJ() error {
	if _, err := exec.LookPath("jj"); err != nil {
		return fmt.Errorf("jj clone requested but `jj` was not found on PATH: %w", err)
	}
	return nil
}

// ansiCSI matches ANSI CSI escape sequences (colors, cursor, clear-line) that
// jj emits around its progress bar.
var ansiCSI = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// jjPercentRe extracts a bare `NN%` from a jj progress chunk (jj emits no phase
// label, unlike git's "Receiving objects: N%").
var jjPercentRe = regexp.MustCompile(`(\d{1,3})%`)

// stripANSI removes ANSI CSI sequences from s.
func stripANSI(s string) string { return ansiCSI.ReplaceAllString(s, "") }

// parseJJProgress extracts a fraction (0..1) from a jj progress chunk after
// stripping ANSI. ok is false when the chunk carries no percentage.
func parseJJProgress(chunk string) (frac float64, ok bool) {
	m := jjPercentRe.FindStringSubmatch(stripANSI(chunk))
	if m == nil {
		return 0, false
	}
	pct, err := strconv.Atoi(m[1])
	if err != nil || pct > 100 {
		return 0, false
	}
	return float64(pct) / 100, true
}

// execJJClone clones url into dir with `jj git clone --colocate`. When emit is
// nil it runs plainly (piped, no progress). When emit is non-nil it runs the
// clone attached to a pseudo-terminal (jj only draws its percentage bar on a
// TTY and has no --progress flag), strips ANSI, and parses `NN%` into progress
// events, falling back silently on unparseable chunks. It is the default
// Engine.JJClone.
func execJJClone(ctx context.Context, url, dir string, emit func(frac float64, phase string)) error {
	if err := ensureJJ(); err != nil {
		return err
	}
	if parent := filepath.Dir(dir); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("creating parent dir %s: %w", parent, err)
		}
	}

	if emit == nil {
		cmd := exec.CommandContext(ctx, "jj", "git", "clone", "--colocate", url, dir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("jj git clone %s: %w: %s", url, err, out)
		}
		return nil
	}

	// Progress path: run under a pty so jj emits its bar; parse NN%.
	cmd := exec.CommandContext(ctx, "jj", "git", "clone", "--colocate", "--color", "never", url, dir)
	f, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("starting jj under pty: %w", err)
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Split(scanCRorLF) // reuse the git progress splitter (\r or \n)
	for sc.Scan() {
		if frac, ok := parseJJProgress(sc.Text()); ok {
			emit(frac, "cloning")
		}
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("jj git clone %s: %w", url, err)
	}
	return nil
}
