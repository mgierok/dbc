package engine

import (
	"errors"
	"testing"
)

func TestWithRollbackError_ReturnsCauseWhenRollbackSucceeds(t *testing.T) {
	// Arrange
	cause := errors.New("update failed")

	// Act
	err := withRollbackError(cause, func() error {
		return nil
	})

	// Assert
	if !errors.Is(err, cause) {
		t.Fatalf("expected error to include cause, got %v", err)
	}
}

func TestWithRollbackError_JoinsCauseAndRollbackError(t *testing.T) {
	// Arrange
	cause := errors.New("update failed")
	rollbackErr := errors.New("rollback failed")

	// Act
	err := withRollbackError(cause, func() error {
		return rollbackErr
	})

	// Assert
	if !errors.Is(err, cause) {
		t.Fatalf("expected error to include cause, got %v", err)
	}
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected error to include rollback error, got %v", err)
	}
}
