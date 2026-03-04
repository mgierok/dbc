package service

import (
	"bytes"
	"errors"
	"testing"
)

func TestInputSpecForType_Boolean(t *testing.T) {
	// Arrange
	spec := InputSpecForType("BOOLEAN")

	// Assert
	if spec.Kind != InputSelect {
		t.Fatalf("expected select input, got %v", spec.Kind)
	}
	expected := []string{"true", "false"}
	if len(spec.Options) != len(expected) {
		t.Fatalf("expected %v options, got %v", expected, spec.Options)
	}
	for i, option := range expected {
		if spec.Options[i] != option {
			t.Fatalf("expected option %q, got %q", option, spec.Options[i])
		}
	}
}

func TestInputSpecForType_Enum(t *testing.T) {
	// Arrange
	spec := InputSpecForType("ENUM('small','medium','large')")

	// Assert
	if spec.Kind != InputSelect {
		t.Fatalf("expected select input, got %v", spec.Kind)
	}
	expected := []string{"small", "medium", "large"}
	if len(spec.Options) != len(expected) {
		t.Fatalf("expected %v options, got %v", expected, spec.Options)
	}
	for i, option := range expected {
		if spec.Options[i] != option {
			t.Fatalf("expected option %q, got %q", option, spec.Options[i])
		}
	}
}

func TestParseValue_AllowsNullWhenNullable(t *testing.T) {
	// Act
	value, err := ParseValue("TEXT", "", true, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !value.IsNull {
		t.Fatalf("expected value to be null")
	}
}

func TestParseValue_RejectsNullWhenNotNullable(t *testing.T) {
	// Act
	_, err := ParseValue("TEXT", "", true, false)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNullNotAllowed) {
		t.Fatalf("expected ErrNullNotAllowed, got %v", err)
	}
}

func TestParseValue_Integer(t *testing.T) {
	// Act
	value, err := ParseValue("INTEGER", "42", false, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	typed, ok := value.Raw.(int64)
	if !ok {
		t.Fatalf("expected int64 raw value, got %T", value.Raw)
	}
	if typed != 42 {
		t.Fatalf("expected 42, got %d", typed)
	}
}

func TestParseValue_Real(t *testing.T) {
	// Act
	value, err := ParseValue("REAL", "3.14", false, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	typed, ok := value.Raw.(float64)
	if !ok {
		t.Fatalf("expected float64 raw value, got %T", value.Raw)
	}
	if typed != 3.14 {
		t.Fatalf("expected 3.14, got %v", typed)
	}
}

func TestParseValue_DoublePrecision(t *testing.T) {
	// Act
	value, err := ParseValue("DOUBLE PRECISION", "2.5", false, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	typed, ok := value.Raw.(float64)
	if !ok {
		t.Fatalf("expected float64 raw value, got %T", value.Raw)
	}
	if typed != 2.5 {
		t.Fatalf("expected 2.5, got %v", typed)
	}
}

func TestParseValue_BlobHex(t *testing.T) {
	// Act
	value, err := ParseValue("BLOB", "0xFF00", false, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	typed, ok := value.Raw.([]byte)
	if !ok {
		t.Fatalf("expected []byte raw value, got %T", value.Raw)
	}
	if !bytes.Equal(typed, []byte{0xff, 0x00}) {
		t.Fatalf("expected blob value %v, got %v", []byte{0xff, 0x00}, typed)
	}
}

func TestParseValue_BooleanFalse(t *testing.T) {
	// Act
	value, err := ParseValue("BOOLEAN", "false", false, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	typed, ok := value.Raw.(int64)
	if !ok {
		t.Fatalf("expected int64 raw value, got %T", value.Raw)
	}
	if typed != 0 {
		t.Fatalf("expected normalized false value 0, got %d", typed)
	}
	if value.Text != "false" {
		t.Fatalf("expected normalized text false, got %q", value.Text)
	}
}

func TestParseValue_RejectsInvalidBoolean(t *testing.T) {
	// Act
	_, err := ParseValue("BOOLEAN", "nope", false, true)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestParseValue_RejectsInvalidInteger(t *testing.T) {
	// Act
	_, err := ParseValue("INTEGER", "abc", false, true)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestParseValue_RejectsInvalidReal(t *testing.T) {
	// Act
	_, err := ParseValue("REAL", "abc", false, true)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestParseValue_RejectsInvalidNumeric(t *testing.T) {
	// Act
	_, err := ParseValue("NUMERIC", "12x", false, true)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestParseValue_RejectsInvalidHexBlob(t *testing.T) {
	// Act
	_, err := ParseValue("BLOB", "0xGG", false, true)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestParseValue_BlobFallsBackToRawBytesForNonHexInput(t *testing.T) {
	// Act
	value, err := ParseValue("BLOB", "FF00", false, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	typed, ok := value.Raw.([]byte)
	if !ok {
		t.Fatalf("expected []byte raw value, got %T", value.Raw)
	}
	if !bytes.Equal(typed, []byte("FF00")) {
		t.Fatalf("expected blob fallback bytes %v, got %v", []byte("FF00"), typed)
	}
}
