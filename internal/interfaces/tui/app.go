package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

var ErrOpenConfigSelector = errors.New("open config selector requested")

func Run(ctx context.Context, listTables *usecase.ListTables, getSchema *usecase.GetSchema, listRecords *usecase.ListRecords, listOperators *usecase.ListOperators, saveChanges *usecase.SaveTableChanges) error {
	model := NewModel(ctx, listTables, getSchema, listRecords, listOperators, saveChanges)
	program := tea.NewProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return err
	}
	runtimeModel, ok := final.(*Model)
	if !ok {
		return errors.New("unexpected runtime model type")
	}
	if runtimeModel.ShouldOpenConfigSelector() {
		return ErrOpenConfigSelector
	}
	return nil
}
