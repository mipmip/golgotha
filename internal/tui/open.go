package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// openURL opens url in the platform's default browser. It is a var so tests can
// stub it and never launch a real browser.
var openURL = func(ctx context.Context, url string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name = "open"
		args = []string{url}
	case "windows":
		name = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default: // linux and other unix
		name = "xdg-open"
		args = []string{url}
	}
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("cannot open browser: %s not found: %w", name, err)
	}
	return exec.CommandContext(ctx, name, args...).Start()
}
