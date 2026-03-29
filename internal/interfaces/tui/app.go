package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/usecase"
)

type RuntimeRunDeps struct {
	ListTables             *usecase.ListTables
	GetSchema              *usecase.GetSchema
	ListRecords            *usecase.ListRecords
	ListOperators          *usecase.ListOperators
	SaveChanges            *usecase.SaveTableChanges
	SaveWorkflow           *usecase.RuntimeSaveWorkflow
	NavigationWorkflow     *usecase.RuntimeNavigationWorkflow
	DatabaseTargetResolver *usecase.RuntimeDatabaseTargetResolver
	Translator             *usecase.StagedChangesTranslator
	DatabaseSelector       *RuntimeDatabaseSelectorDeps
	Close                  func()
}

type RuntimeExitAction int

const (
	RuntimeExitActionQuit RuntimeExitAction = iota + 1
	RuntimeExitActionOpenDatabaseNext
)

type RuntimeExitResult struct {
	Action       RuntimeExitAction
	NextDatabase DatabaseOption
}

func runtimeExitResultQuit() RuntimeExitResult {
	return RuntimeExitResult{Action: RuntimeExitActionQuit}
}

func runtimeExitResultOpenDatabaseNext(selected DatabaseOption) RuntimeExitResult {
	return RuntimeExitResult{
		Action:       RuntimeExitActionOpenDatabaseNext,
		NextDatabase: selected,
	}
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
) (RuntimeExitResult, error) {
	model := NewModel(ctx, runtimeDeps, nil)
	program := newRuntimeProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		model.closeRuntimeResources()
		return RuntimeExitResult{}, err
	}
	runtimeModel, ok := final.(*Model)
	if !ok {
		model.closeRuntimeResources()
		return RuntimeExitResult{}, errors.New("unexpected runtime model type")
	}
	runtimeModel.closeRuntimeResources()
	return runtimeModel.exitResultOrDefault(), nil
}
