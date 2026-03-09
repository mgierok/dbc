package selector

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

type stubSelectorProgram struct {
	run func() (tea.Model, error)
}

func (s stubSelectorProgram) Run() (tea.Model, error) {
	return s.run()
}

type stubUnexpectedSelectorModel struct{}

func (stubUnexpectedSelectorModel) Init() tea.Cmd {
	return nil
}

func (stubUnexpectedSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return stubUnexpectedSelectorModel{}, nil
}

func (stubUnexpectedSelectorModel) View() string {
	return ""
}

func TestSelectDatabaseWithState_ReturnsErrorWhenManagerMissing(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	_, err := SelectDatabaseWithState(ctx, nil, SelectorLaunchState{})

	// Assert
	if err == nil {
		t.Fatal("expected selector manager validation error")
	}
	if err.Error() != "selector manager is required" {
		t.Fatalf("expected selector manager validation error, got %v", err)
	}
}

func TestSelectDatabaseWithState_ReturnsProgramError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("run failed")
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return nil, expectedErr
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	_, err := SelectDatabaseWithState(context.Background(), manager, SelectorLaunchState{})

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSelectDatabaseWithState_ReturnsErrorWhenFinalModelTypeIsUnexpected(t *testing.T) {
	// Arrange
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return stubUnexpectedSelectorModel{}, nil
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	_, err := SelectDatabaseWithState(context.Background(), manager, SelectorLaunchState{})

	// Assert
	if err == nil {
		t.Fatal("expected unexpected selector state error")
	}
	if err.Error() != "unexpected selector state" {
		t.Fatalf("expected unexpected selector state error, got %v", err)
	}
}

func TestSelectDatabaseWithState_ReturnsCanceledError(t *testing.T) {
	// Arrange
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return &databaseSelectorModel{canceled: true}, nil
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	_, err := SelectDatabaseWithState(context.Background(), manager, SelectorLaunchState{})

	// Assert
	if !errors.Is(err, ErrDatabaseSelectionCanceled) {
		t.Fatalf("expected error %v, got %v", ErrDatabaseSelectionCanceled, err)
	}
}

func TestSelectDatabaseWithState_ReturnsUnfinishedErrorWhenSelectionNotConfirmed(t *testing.T) {
	// Arrange
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return &databaseSelectorModel{
					options: []DatabaseOption{
						{Name: "local", ConnString: "/tmp/local.sqlite"},
					},
				}, nil
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	_, err := SelectDatabaseWithState(context.Background(), manager, SelectorLaunchState{})

	// Assert
	if !errors.Is(err, ErrDatabaseSelectionUnfinished) {
		t.Fatalf("expected error %v, got %v", ErrDatabaseSelectionUnfinished, err)
	}
}

func TestSelectDatabaseWithState_ReturnsUnfinishedErrorWhenSelectedIndexIsInvalid(t *testing.T) {
	// Arrange
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return &databaseSelectorModel{
					chosen: true,
					browse: selectorBrowseState{
						selected: 2,
					},
					options: []DatabaseOption{
						{Name: "local", ConnString: "/tmp/local.sqlite"},
					},
				}, nil
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	_, err := SelectDatabaseWithState(context.Background(), manager, SelectorLaunchState{})

	// Assert
	if !errors.Is(err, ErrDatabaseSelectionUnfinished) {
		t.Fatalf("expected error %v, got %v", ErrDatabaseSelectionUnfinished, err)
	}
}

func TestSelectDatabaseWithState_ReturnsSelectedOption(t *testing.T) {
	// Arrange
	expected := DatabaseOption{Name: "analytics", ConnString: "/tmp/analytics.sqlite"}
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return &databaseSelectorModel{
					chosen: true,
					browse: selectorBrowseState{
						selected: 1,
					},
					options: []DatabaseOption{
						{Name: "local", ConnString: "/tmp/local.sqlite"},
						expected,
					},
				}, nil
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
			{Name: "analytics", Path: "/tmp/analytics.sqlite"},
		},
	}

	// Act
	selected, err := SelectDatabaseWithState(context.Background(), manager, SelectorLaunchState{})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selected != expected {
		t.Fatalf("expected selected option %+v, got %+v", expected, selected)
	}
}

func TestSelectDatabase_UsesDefaultLaunchState(t *testing.T) {
	// Arrange
	expected := DatabaseOption{Name: "local", ConnString: "/tmp/local.sqlite"}
	originalFactory := newSelectorProgram
	t.Cleanup(func() {
		newSelectorProgram = originalFactory
	})
	newSelectorProgram = func(model tea.Model, options ...tea.ProgramOption) selectorProgram {
		return stubSelectorProgram{
			run: func() (tea.Model, error) {
				return &databaseSelectorModel{
					chosen: true,
					options: []DatabaseOption{
						expected,
					},
				}, nil
			},
		}
	}
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	selected, err := SelectDatabase(context.Background(), manager)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selected != expected {
		t.Fatalf("expected selected option %+v, got %+v", expected, selected)
	}
}
