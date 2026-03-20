package usecase_test

import (
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestDirtyNavigationPolicy_BuildTableSwitchPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildTableSwitchPrompt(3)

	// Assert
	if prompt.Title != "Switch Table" {
		t.Fatalf("expected title Switch Table, got %q", prompt.Title)
	}
	expectedMessage := "Switching tables will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data?"
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
	}
	if len(prompt.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != usecase.DirtyDecisionDiscard || prompt.Options[0].Label != "Discard changes and switch table" {
		t.Fatalf("unexpected first option: %#v", prompt.Options[0])
	}
	if prompt.Options[1].ID != usecase.DirtyDecisionCancel || prompt.Options[1].Label != "Continue editing" {
		t.Fatalf("unexpected second option: %#v", prompt.Options[1])
	}
}

func TestDirtyNavigationPolicy_BuildTableSwitchPrompt_ClampsNegativeCountToZero(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildTableSwitchPrompt(-5)

	// Assert
	expectedMessage := "Switching tables will cause loss of unsaved data (0 rows). Are you sure you want to discard unsaved data?"
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
	}
}

func TestDirtyNavigationPolicy_BuildConfigPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildConfigPrompt()

	// Assert
	if prompt.Title != "Config" {
		t.Fatalf("expected title Config, got %q", prompt.Title)
	}
	if prompt.Message != "Unsaved changes detected. Choose save, discard, or cancel." {
		t.Fatalf("unexpected message: %q", prompt.Message)
	}
	if len(prompt.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != usecase.DirtyDecisionSave || prompt.Options[0].Label != "Save and open config" {
		t.Fatalf("unexpected save option: %#v", prompt.Options[0])
	}
	if prompt.Options[1].ID != usecase.DirtyDecisionDiscard || prompt.Options[1].Label != "Discard and open config" {
		t.Fatalf("unexpected discard option: %#v", prompt.Options[1])
	}
	if prompt.Options[2].ID != usecase.DirtyDecisionCancel || prompt.Options[2].Label != "Cancel" {
		t.Fatalf("unexpected cancel option: %#v", prompt.Options[2])
	}
}

func TestDirtyNavigationPolicy_BuildDatabaseReloadPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	policy := usecase.NewDirtyNavigationPolicy()

	prompt := policy.BuildDatabaseReloadPrompt(3)

	if prompt.Title != "Reload Database" {
		t.Fatalf("expected title Reload Database, got %q", prompt.Title)
	}
	expectedMessage := "Reloading the current database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel."
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
	}
	if len(prompt.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != usecase.DirtyDecisionSave || prompt.Options[0].Label != "Save and reload database" {
		t.Fatalf("unexpected save option: %#v", prompt.Options[0])
	}
	if prompt.Options[1].ID != usecase.DirtyDecisionDiscard || prompt.Options[1].Label != "Discard changes and reload database" {
		t.Fatalf("unexpected discard option: %#v", prompt.Options[1])
	}
	if prompt.Options[2].ID != usecase.DirtyDecisionCancel || prompt.Options[2].Label != "Cancel" {
		t.Fatalf("unexpected cancel option: %#v", prompt.Options[2])
	}
}

func TestDirtyNavigationPolicy_BuildDatabaseOpenPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	policy := usecase.NewDirtyNavigationPolicy()

	prompt := policy.BuildDatabaseOpenPrompt(3)

	if prompt.Title != "Open Database" {
		t.Fatalf("expected title Open Database, got %q", prompt.Title)
	}
	expectedMessage := "Opening another database will cause loss of unsaved data (3 rows) unless you save first. Choose save, discard, or cancel."
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
	}
	if len(prompt.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != usecase.DirtyDecisionSave || prompt.Options[0].Label != "Save and open database" {
		t.Fatalf("unexpected save option: %#v", prompt.Options[0])
	}
	if prompt.Options[1].ID != usecase.DirtyDecisionDiscard || prompt.Options[1].Label != "Discard changes and open database" {
		t.Fatalf("unexpected discard option: %#v", prompt.Options[1])
	}
	if prompt.Options[2].ID != usecase.DirtyDecisionCancel || prompt.Options[2].Label != "Cancel" {
		t.Fatalf("unexpected cancel option: %#v", prompt.Options[2])
	}
}

func TestDirtyNavigationPolicy_BuildQuitPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildQuitPrompt(3)

	// Assert
	if prompt.Title != "Quit" {
		t.Fatalf("expected title Quit, got %q", prompt.Title)
	}
	expectedMessage := "Quitting will cause loss of unsaved data (3 rows). Are you sure you want to discard unsaved data and quit?"
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
	}
	if len(prompt.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != usecase.DirtyDecisionDiscard || prompt.Options[0].Label != "Discard changes and quit" {
		t.Fatalf("unexpected first option: %#v", prompt.Options[0])
	}
	if prompt.Options[1].ID != usecase.DirtyDecisionCancel || prompt.Options[1].Label != "Continue editing" {
		t.Fatalf("unexpected second option: %#v", prompt.Options[1])
	}
}

func TestDirtyNavigationPolicy_BuildQuitPrompt_ClampsNegativeCountToZero(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildQuitPrompt(-5)

	// Assert
	expectedMessage := "Quitting will cause loss of unsaved data (0 rows). Are you sure you want to discard unsaved data and quit?"
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
	}
}
