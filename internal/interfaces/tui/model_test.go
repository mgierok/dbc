package tui

import "testing"

func TestNormalizePastedValue_StripsBrackets(t *testing.T) {
	// Arrange
	input := "[hello]"

	// Act
	result := normalizePastedValue(input)

	// Assert
	if result != "hello" {
		t.Fatalf("expected %q, got %q", "hello", result)
	}
}

func TestNormalizePastedValue_LeavesPlainValue(t *testing.T) {
	// Arrange
	input := "hello"

	// Act
	result := normalizePastedValue(input)

	// Assert
	if result != "hello" {
		t.Fatalf("expected %q, got %q", "hello", result)
	}
}

func TestInsertAtCursor_AddsAtPosition(t *testing.T) {
	// Arrange
	value := "abcd"

	// Act
	result, cursor := insertAtCursor(value, "X", 2)

	// Assert
	if result != "abXcd" {
		t.Fatalf("expected %q, got %q", "abXcd", result)
	}
	if cursor != 3 {
		t.Fatalf("expected cursor 3, got %d", cursor)
	}
}

func TestDeleteAtCursor_RemovesPreviousRune(t *testing.T) {
	// Arrange
	value := "abcd"

	// Act
	result, cursor := deleteAtCursor(value, 3)

	// Assert
	if result != "abd" {
		t.Fatalf("expected %q, got %q", "abd", result)
	}
	if cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", cursor)
	}
}
