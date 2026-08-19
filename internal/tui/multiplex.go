package tui

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/mipmip/huphop/internal/clonepath"
)

// runArgv executes argv without a shell, wired to the user's terminal. It is the
// default Model.runCommand.
func runArgv(argv []string) error {
	if len(argv) == 0 {
		return errors.New("empty command")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// splitArgs splits a rendered command string into argv using POSIX-ish
// shell-quoting rules (single quotes, double quotes, backslash escapes) WITHOUT
// invoking a shell, so repository or owner names cannot inject shell behavior.
func splitArgs(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inArg := false
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\'':
			inArg = true
			i++
			for i < len(s) && s[i] != '\'' {
				cur.WriteByte(s[i])
				i++
			}
			if i >= len(s) {
				return nil, errors.New("unterminated single quote")
			}
			i++ // consume closing quote
		case c == '"':
			inArg = true
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' && i+1 < len(s) {
					if n := s[i+1]; n == '"' || n == '\\' || n == '$' || n == '`' {
						cur.WriteByte(n)
						i += 2
						continue
					}
				}
				cur.WriteByte(s[i])
				i++
			}
			if i >= len(s) {
				return nil, errors.New("unterminated double quote")
			}
			i++ // consume closing quote
		case c == '\\':
			if i+1 >= len(s) {
				return nil, errors.New("trailing backslash")
			}
			cur.WriteByte(s[i+1])
			i += 2
			inArg = true
		case c == ' ' || c == '\t' || c == '\n':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
			i++
		default:
			cur.WriteByte(c)
			inArg = true
			i++
		}
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}

// switchData is the template context for a mode's switch_command: the clone-path
// fields plus the resolved local Target path.
type switchData struct {
	clonepath.Data
	Target string
}

// renderSwitchCommand renders a switch_command template for a repo item.
func (m *Model) renderSwitchCommand(tpl string, it repoItem) (string, error) {
	data := switchData{
		Data:   clonepath.NewData(m.cfg.BaseDir, it.Provider, it.Provider.WebURL, it.Repo.Owner, it.Repo.Name),
		Target: it.Target,
	}
	t, err := template.New("switch").Option("missingkey=error").Parse(tpl)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// multiplexActive reports whether the active mode is a multiplex-style mode
// (its primary action is clone-then-switch, so multiselect is irrelevant).
func (m *Model) multiplexActive() bool {
	return m.activeSwitchCommand() != ""
}

// activeSwitchCommand returns the switch_command for the active mode, or "".
func (m *Model) activeSwitchCommand() string {
	if m.cfg == nil {
		return ""
	}
	if mc, ok := m.cfg.Modes[m.mode]; ok {
		return mc.SwitchCommand
	}
	return ""
}

// runSwitch renders, splits and executes the switch_command for a repo item.
func (m *Model) runSwitch(tpl string, it repoItem) error {
	rendered, err := m.renderSwitchCommand(tpl, it)
	if err != nil {
		return err
	}
	argv, err := splitArgs(rendered)
	if err != nil {
		return err
	}
	if len(argv) == 0 {
		return errors.New("switch_command rendered to an empty command")
	}
	run := m.runCommand
	if run == nil {
		run = runArgv
	}
	return run(argv)
}
