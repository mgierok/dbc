package engine

import (
	"errors"
	"testing"

	"github.com/mgierok/dbc/internal/domain/model"
)

func TestBuildFilterClause_NoFilter(t *testing.T) {
	// Arrange
	var filter *model.Filter

	// Act
	clause, args, err := buildFilterClause(filter)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if clause != "" {
		t.Fatalf("expected empty clause, got %q", clause)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestBuildFilterClause_WithValueOperator(t *testing.T) {
	// Arrange
	filter := &model.Filter{
		Column: "name",
		Operator: model.Operator{
			SQL:           "=",
			RequiresValue: true,
		},
		Value: "alice",
	}

	// Act
	clause, args, err := buildFilterClause(filter)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expectedClause := `WHERE "name" = ?`
	if clause != expectedClause {
		t.Fatalf("expected clause %q, got %q", expectedClause, clause)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("expected args [alice], got %v", args)
	}
}

func TestBuildFilterClause_IsNullOperator(t *testing.T) {
	// Arrange
	filter := &model.Filter{
		Column: "deleted_at",
		Operator: model.Operator{
			SQL:           "IS NULL",
			RequiresValue: false,
		},
	}

	// Act
	clause, args, err := buildFilterClause(filter)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expectedClause := `WHERE "deleted_at" IS NULL`
	if clause != expectedClause {
		t.Fatalf("expected clause %q, got %q", expectedClause, clause)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestBuildFilterClause_UnknownOperator(t *testing.T) {
	// Arrange
	filter := &model.Filter{
		Column: "name",
		Operator: model.Operator{
			SQL:           "DROP TABLE",
			RequiresValue: true,
		},
		Value: "x",
	}

	// Act
	_, _, err := buildFilterClause(filter)

	// Assert
	if !errors.Is(err, ErrUnknownOperator) {
		t.Fatalf("expected error %v, got %v", ErrUnknownOperator, err)
	}
}

func TestBuildFilterClause_MissingColumn(t *testing.T) {
	// Arrange
	filter := &model.Filter{
		Column: " ",
		Operator: model.Operator{
			SQL:           "=",
			RequiresValue: true,
		},
		Value: "x",
	}

	// Act
	_, _, err := buildFilterClause(filter)

	// Assert
	if !errors.Is(err, ErrMissingFilterColumn) {
		t.Fatalf("expected error %v, got %v", ErrMissingFilterColumn, err)
	}
}
