package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

var ErrOpenConfigSelector = errors.New("open config selector requested")

type runtimeProgram interface {
	Run() (tea.Model, error)
}

var newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
	return tea.NewProgram(model, options...)
}

func Run(
	ctx context.Context,
	listTables *usecase.ListTables,
	getSchema *usecase.GetSchema,
	listRecords *usecase.ListRecords,
	listOperators *usecase.ListOperators,
	saveChanges *usecase.SaveTableChanges,
	translator *usecase.StagedChangesTranslator,
) error {
	model := NewModel(ctx, listTables, getSchema, listRecords, listOperators, saveChanges, translator)
	program := newRuntimeProgram(model, tea.WithAltScreen())
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
