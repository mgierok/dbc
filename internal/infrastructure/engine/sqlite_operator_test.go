package engine

import (
	"testing"

	"github.com/mgierok/dbc/internal/domain/model"
)

func TestOperatorsForType_ReturnsSQLiteOperators(t *testing.T) {
	// Arrange
	columnType := "INTEGER"

	// Act
	operators := operatorsForType(columnType)

	// Assert
	if len(operators) == 0 {
		t.Fatal("expected operators, got none")
	}
	foundIsNull := false
	for _, operator := range operators {
		if operator.Kind == model.OperatorKindIsNull {
			foundIsNull = true
			break
		}
	}
	if !foundIsNull {
		t.Fatal("expected IS NULL operator")
	}
}
