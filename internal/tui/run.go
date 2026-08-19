package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/config"
)

// Run builds the model from cfg and runs the Bubble Tea program until the user
// quits. It is the entrypoint used by the `hup tui` command. When mode is
// non-empty it overrides the configured default_mode.
func Run(cfg *config.Config, mode string) error {
	m := New(cfg)
	if mode != "" {
		if err := m.SetMode(mode); err != nil {
			return err
		}
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
