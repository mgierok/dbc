package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func Run(ctx context.Context, listTables *usecase.ListTables, getSchema *usecase.GetSchema, listRecords *usecase.ListRecords, listOperators *usecase.ListOperators) error {
	model := NewModel(ctx, listTables, getSchema, listRecords, listOperators)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
