package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

type RuntimeRunDeps struct {
	ListTables       *usecase.ListTables
	GetSchema        *usecase.GetSchema
	ListRecords      *usecase.ListRecords
	ListOperators    *usecase.ListOperators
	SaveChanges      *usecase.SaveTableChanges
	Translator       *usecase.StagedChangesTranslator
	DatabaseSelector *RuntimeDatabaseSelectorDeps
	Close            func()
}

type RuntimeDatabaseSwitcher interface {
	Switch(ctx context.Context, selected DatabaseOption) (RuntimeRunDeps, error)
}

type RuntimeDatabaseSwitchFunc func(ctx context.Context, selected DatabaseOption) (RuntimeRunDeps, error)

func (f RuntimeDatabaseSwitchFunc) Switch(ctx context.Context, selected DatabaseOption) (RuntimeRunDeps, error) {
	return f(ctx, selected)
}

type runtimeProgram interface {
	Run() (tea.Model, error)
}

var newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
	return tea.NewProgram(model, options...)
}

func Run(
	ctx context.Context,
	runtimeDeps RuntimeRunDeps,
	runtimeSession *RuntimeSessionState,
) error {
	model := NewModel(ctx, runtimeDeps, runtimeSession)
	program := newRuntimeProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		model.closeRuntimeResources()
		return err
	}
	runtimeModel, ok := final.(*Model)
	if !ok {
		model.closeRuntimeResources()
		return errors.New("unexpected runtime model type")
	}
	runtimeModel.closeRuntimeResources()
	return nil
}
