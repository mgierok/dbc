package usecase_test

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestRuntimeRecordLimitPolicy_Default_ReturnsExpectedLimit(t *testing.T) {
	// Arrange
	policy := usecase.NewRuntimeRecordLimitPolicy()

	// Act
	limit := policy.Default()

	// Assert
	if limit != 20 {
		t.Fatalf("expected default record limit 20, got %d", limit)
	}
}

func TestRuntimeRecordLimitPolicy_Validate_AcceptsBoundaryValues(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{name: "min", limit: 1},
		{name: "max", limit: 1000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			policy := usecase.NewRuntimeRecordLimitPolicy()

			// Act
			err := policy.Validate(tc.limit)

			// Assert
			if err != nil {
				t.Fatalf("expected limit %d to be accepted, got error %v", tc.limit, err)
			}
		})
	}
}

func TestRuntimeRecordLimitPolicy_Validate_RejectsOutOfRangeValuesWithDeterministicHint(t *testing.T) {
	tests := []struct {
		name  string
		limit int
	}{
		{name: "zero", limit: 0},
		{name: "negative", limit: -1},
		{name: "too large", limit: 1001},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			policy := usecase.NewRuntimeRecordLimitPolicy()

			// Act
			err := policy.Validate(tc.limit)

			// Assert
			if err == nil {
				t.Fatalf("expected limit %d to be rejected", tc.limit)
			}
			if !strings.Contains(err.Error(), "expected :set limit=<1-1000>") {
				t.Fatalf("expected deterministic hint for limit %d, got %v", tc.limit, err)
			}
		})
	}
}

func TestRuntimeRecordLimitPolicy_Effective_ReturnsDefaultForUnsetValue(t *testing.T) {
	// Arrange
	policy := usecase.NewRuntimeRecordLimitPolicy()

	// Act
	limit := policy.Effective(0)

	// Assert
	if limit != 20 {
		t.Fatalf("expected default effective limit 20, got %d", limit)
	}
}

func TestRuntimeRecordLimitPolicy_Effective_ClampsOversizedValue(t *testing.T) {
	// Arrange
	policy := usecase.NewRuntimeRecordLimitPolicy()

	// Act
	limit := policy.Effective(1001)

	// Assert
	if limit != 1000 {
		t.Fatalf("expected oversized effective limit to clamp to 1000, got %d", limit)
	}
}
