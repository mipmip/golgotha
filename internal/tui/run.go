package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/golgotha/internal/config"
)

// Run builds the model from cfg and runs the Bubble Tea program until the user
// quits. It is the entrypoint used by the `gol tui` command.
func Run(cfg *config.Config) error {
	m := New(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
