package usecase_test

import (
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestDirtyNavigationPolicy_BuildPrompt_ReturnsExpectedCopyAndOptions(t *testing.T) {
	t.Parallel()

	policy := usecase.NewDirtyNavigationPolicy()

	tests := []struct {
		name     string
		build    func() usecase.DirtyDecisionPrompt
		expected usecase.DirtyDecisionPrompt
	}{
		{
			name:  "table switch",
			build: func() usecase.DirtyDecisionPrompt { return policy.BuildTableSwitchPrompt(3) },
			expected: usecase.DirtyDecisionPrompt{
				Title:   "Switch Table",
				Message: "Switching tables will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data?",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and switch table"},
					{ID: usecase.DirtyDecisionCancel, Label: "Continue editing"},
				},
			},
		},
		{
			name:  "config",
			build: func() usecase.DirtyDecisionPrompt { return policy.BuildConfigPrompt() },
			expected: usecase.DirtyDecisionPrompt{
				Title:   "Config",
				Message: "Unsaved changes detected. Choose save, discard, or cancel.",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionSave, Label: "Save and open config"},
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard and open config"},
					{ID: usecase.DirtyDecisionCancel, Label: "Cancel"},
				},
			},
		},
		{
			name:  "database reload",
			build: func() usecase.DirtyDecisionPrompt { return policy.BuildDatabaseReloadPrompt(3) },
			expected: usecase.DirtyDecisionPrompt{
				Title:   "Reload Database",
				Message: "Reloading the current database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel.",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionSave, Label: "Save and reload database"},
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and reload database"},
					{ID: usecase.DirtyDecisionCancel, Label: "Cancel"},
				},
			},
		},
		{
			name:  "database open",
			build: func() usecase.DirtyDecisionPrompt { return policy.BuildDatabaseOpenPrompt(3) },
			expected: usecase.DirtyDecisionPrompt{
				Title:   "Open Database",
				Message: "Opening another database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel.",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionSave, Label: "Save and open database"},
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and open database"},
					{ID: usecase.DirtyDecisionCancel, Label: "Cancel"},
				},
			},
		},
		{
			name:  "quit",
			build: func() usecase.DirtyDecisionPrompt { return policy.BuildQuitPrompt(3) },
			expected: usecase.DirtyDecisionPrompt{
				Title:   "Quit",
				Message: "Quitting will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data and quit?",
				Options: []usecase.DirtyDecisionOption{
					{ID: usecase.DirtyDecisionDiscard, Label: "Discard changes and quit"},
					{ID: usecase.DirtyDecisionCancel, Label: "Continue editing"},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prompt := tc.build()

			assertDirtyDecisionPrompt(t, prompt, tc.expected)
		})
	}
}

func TestDirtyNavigationPolicy_BuildPrompt_ClampsNegativeCountToZero(t *testing.T) {
	t.Parallel()

	policy := usecase.NewDirtyNavigationPolicy()

	tests := []struct {
		name            string
		build           func() usecase.DirtyDecisionPrompt
		expectedMessage string
	}{
		{
			name:            "table switch",
			build:           func() usecase.DirtyDecisionPrompt { return policy.BuildTableSwitchPrompt(-5) },
			expectedMessage: "Switching tables will cause loss of unsaved data (0 rows). Are you sure you want to discard unsaved data?",
		},
		{
			name:            "database reload",
			build:           func() usecase.DirtyDecisionPrompt { return policy.BuildDatabaseReloadPrompt(-5) },
			expectedMessage: "Reloading the current database will cause loss of unsaved data (0 rows) unless you save first. Choose save, discard, or cancel.",
		},
		{
			name:            "database open",
			build:           func() usecase.DirtyDecisionPrompt { return policy.BuildDatabaseOpenPrompt(-5) },
			expectedMessage: "Opening another database will cause loss of unsaved data (0 rows) unless you save first. Choose save, discard, or cancel.",
		},
		{
			name:            "quit",
			build:           func() usecase.DirtyDecisionPrompt { return policy.BuildQuitPrompt(-5) },
			expectedMessage: "Quitting will cause loss of unsaved data (0 rows). Are you sure you want to discard unsaved data and quit?",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prompt := tc.build()

			if prompt.Message != tc.expectedMessage {
				t.Fatalf("expected message %q, got %q", tc.expectedMessage, prompt.Message)
			}
		})
	}
}

func assertDirtyDecisionPrompt(t *testing.T, actual, expected usecase.DirtyDecisionPrompt) {
	t.Helper()

	if actual.Title != expected.Title {
		t.Fatalf("expected title %q, got %q", expected.Title, actual.Title)
	}
	if actual.Message != expected.Message {
		t.Fatalf("expected message %q, got %q", expected.Message, actual.Message)
	}
	if !reflect.DeepEqual(actual.Options, expected.Options) {
		t.Fatalf("expected options %#v, got %#v", expected.Options, actual.Options)
	}
}
