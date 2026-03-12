package usecase_test

import (
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestDirtyNavigationPolicy_BuildConfigPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildConfigPrompt(3)

	// Assert
	if prompt.Title != "Config" {
		t.Fatalf("expected title Config, got %q", prompt.Title)
	}
	expectedMessage := "Unsaved changes detected in 3 tables. Save all changes before you open config, discard them, or cancel."
	if prompt.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, prompt.Message)
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

func TestDirtyNavigationPolicy_BuildQuitPrompt_UsesExpectedCopyAndOptions(t *testing.T) {
	// Arrange
	policy := usecase.NewDirtyNavigationPolicy()

	// Act
	prompt := policy.BuildQuitPrompt(1)

	// Assert
	if prompt.Title != "Quit" {
		t.Fatalf("expected title Quit, got %q", prompt.Title)
	}
	if prompt.Message != "Unsaved changes detected. Save changes before you quit, discard them, or cancel." {
		t.Fatalf("unexpected message: %q", prompt.Message)
	}
	if len(prompt.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != usecase.DirtyDecisionSave || prompt.Options[0].Label != "Save and quit" {
		t.Fatalf("unexpected save option: %#v", prompt.Options[0])
	}
	if prompt.Options[1].ID != usecase.DirtyDecisionDiscard || prompt.Options[1].Label != "Discard and quit" {
		t.Fatalf("unexpected discard option: %#v", prompt.Options[1])
	}
	if prompt.Options[2].ID != usecase.DirtyDecisionCancel || prompt.Options[2].Label != "Cancel" {
		t.Fatalf("unexpected cancel option: %#v", prompt.Options[2])
	}
}
