package engine

import "testing"

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
		if operator.SQL == "IS NULL" {
			foundIsNull = true
			break
		}
	}
	if !foundIsNull {
		t.Fatal("expected IS NULL operator")
	}
}
