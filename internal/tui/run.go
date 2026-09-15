package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run carrega o arquivo SQL em path e inicia a interface interativa de
// terminal para navegação pelos comandos INSERT.
func Run(path string) error {
	model, err := New(path)
	if err != nil {
		return fmt.Errorf("carregar arquivo para visualização: %w", err)
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("executar interface de visualização: %w", err)
	}

	return nil
}
