package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type stubRuntimeProgram struct {
	run func() (tea.Model, error)
}

func (s stubRuntimeProgram) Run() (tea.Model, error) {
	return s.run()
}

type stubFinalModel struct{}

func (stubFinalModel) Init() tea.Cmd {
	return nil
}

func (stubFinalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return stubFinalModel{}, nil
}

func (stubFinalModel) View() string {
	return ""
}

func TestRun_ReturnsProgramError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("run failed")
	closed := false
	originalFactory := newRuntimeProgram
	t.Cleanup(func() {
		newRuntimeProgram = originalFactory
	})
	newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
		return stubRuntimeProgram{
			run: func() (tea.Model, error) {
				return nil, expectedErr
			},
		}
	}

	// Act
	_, err := Run(context.Background(), RuntimeRunDeps{
		Close: func() {
			closed = true
		},
	})

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if !closed {
		t.Fatal("expected runtime resources to close on program error")
	}
}

func TestRun_ReturnsErrorWhenFinalModelTypeIsUnexpected(t *testing.T) {
	// Arrange
	closed := false
	originalFactory := newRuntimeProgram
	t.Cleanup(func() {
		newRuntimeProgram = originalFactory
	})
	newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
		return stubRuntimeProgram{
			run: func() (tea.Model, error) {
				return stubFinalModel{}, nil
			},
		}
	}

	// Act
	_, err := Run(context.Background(), RuntimeRunDeps{
		Close: func() {
			closed = true
		},
	})

	// Assert
	if err == nil {
		t.Fatal("expected unexpected model type error")
	}
	if err.Error() != "unexpected runtime model type" {
		t.Fatalf("expected unexpected model type error, got %v", err)
	}
	if !closed {
		t.Fatal("expected runtime resources to close when final model type is unexpected")
	}
}

func TestRun_ClosesFinalRuntimeResourcesOnNormalCompletion(t *testing.T) {
	// Arrange
	closed := false
	originalFactory := newRuntimeProgram
	t.Cleanup(func() {
		newRuntimeProgram = originalFactory
	})
	newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
		return stubRuntimeProgram{
			run: func() (tea.Model, error) {
				return &Model{
					runtimeClose: func() {
						closed = true
					},
				}, nil
			},
		}
	}

	// Act
	result, err := Run(context.Background(), RuntimeRunDeps{})

	// Assert
	if err != nil {
		t.Fatalf("expected no run error, got %v", err)
	}
	if result.Action != RuntimeExitActionQuit {
		t.Fatalf("expected normal completion to return quit action, got %v", result.Action)
	}
	if !closed {
		t.Fatal("expected final runtime resources to close on normal completion")
	}
}

func TestRun_ReturnsNilOnNormalCompletion(t *testing.T) {
	// Arrange
	originalFactory := newRuntimeProgram
	t.Cleanup(func() {
		newRuntimeProgram = originalFactory
	})
	newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
		return stubRuntimeProgram{
			run: func() (tea.Model, error) {
				return &Model{}, nil
			},
		}
	}

	// Act
	result, err := Run(context.Background(), RuntimeRunDeps{})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Action != RuntimeExitActionQuit {
		t.Fatalf("expected normal completion to return quit action, got %v", result.Action)
	}
}
