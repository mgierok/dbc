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
	_, err := Run(context.Background(), nil, nil, nil, nil, nil, nil, nil, nil)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestRun_ReturnsErrorWhenFinalModelTypeIsUnexpected(t *testing.T) {
	// Arrange
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
	_, err := Run(context.Background(), nil, nil, nil, nil, nil, nil, nil, nil)

	// Assert
	if err == nil {
		t.Fatal("expected unexpected model type error")
	}
	if err.Error() != "unexpected runtime model type" {
		t.Fatalf("expected unexpected model type error, got %v", err)
	}
}

func TestRun_ReturnsRuntimeExitRequestWhenModelRequestsDatabaseSwitch(t *testing.T) {
	// Arrange
	originalFactory := newRuntimeProgram
	t.Cleanup(func() {
		newRuntimeProgram = originalFactory
	})
	newRuntimeProgram = func(model tea.Model, options ...tea.ProgramOption) runtimeProgram {
		return stubRuntimeProgram{
			run: func() (tea.Model, error) {
				return &Model{ui: runtimeUIState{
					exitRequest: RuntimeExitRequest{
						Action:   RuntimeExitActionSwitchDatabase,
						Database: DatabaseOption{ConnString: "/tmp/analytics.sqlite"},
					},
				}}, nil
			},
		}
	}

	// Act
	exitRequest, err := Run(context.Background(), nil, nil, nil, nil, nil, nil, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no run error, got %v", err)
	}
	if exitRequest.Action != RuntimeExitActionSwitchDatabase {
		t.Fatalf("expected switch-database exit request, got %+v", exitRequest)
	}
	if exitRequest.Database.ConnString != "/tmp/analytics.sqlite" {
		t.Fatalf("expected switch target /tmp/analytics.sqlite, got %q", exitRequest.Database.ConnString)
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
	exitRequest, err := Run(context.Background(), nil, nil, nil, nil, nil, nil, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exitRequest.Action != RuntimeExitActionNone {
		t.Fatalf("expected no runtime exit request, got %+v", exitRequest)
	}
}
