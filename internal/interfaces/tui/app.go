package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

type RuntimeExitAction int

const (
	RuntimeExitActionNone RuntimeExitAction = iota
	RuntimeExitActionSwitchDatabase
)

type RuntimeExitRequest struct {
	Action   RuntimeExitAction
	Database DatabaseOption
}

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
	runtimeSession *RuntimeSessionState,
	runtimeDatabaseSelectorDeps *RuntimeDatabaseSelectorDeps,
) (RuntimeExitRequest, error) {
	model := NewModel(ctx, listTables, getSchema, listRecords, listOperators, saveChanges, translator, runtimeSession, runtimeDatabaseSelectorDeps)
	program := newRuntimeProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return RuntimeExitRequest{}, err
	}
	runtimeModel, ok := final.(*Model)
	if !ok {
		return RuntimeExitRequest{}, errors.New("unexpected runtime model type")
	}
	return runtimeModel.RuntimeExitRequest(), nil
}
