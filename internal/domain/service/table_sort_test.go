package service_test

import (
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/domain/model"
	"github.com/mgierok/dbc/internal/domain/service"
)

func TestSortedTablesByName_ReturnsAlphabeticalOrder(t *testing.T) {
	// Arrange
	tables := []model.Table{
		{Name: "users"},
		{Name: "accounts"},
		{Name: "orders"},
	}

	// Act
	sorted := service.SortedTablesByName(tables)

	// Assert
	expected := []model.Table{
		{Name: "accounts"},
		{Name: "orders"},
		{Name: "users"},
	}
	if !reflect.DeepEqual(sorted, expected) {
		t.Fatalf("expected %v, got %v", expected, sorted)
	}
}

func TestSortedTablesByName_DoesNotMutateInput(t *testing.T) {
	// Arrange
	tables := []model.Table{
		{Name: "b"},
		{Name: "a"},
	}
	original := append([]model.Table(nil), tables...)

	// Act
	_ = service.SortedTablesByName(tables)

	// Assert
	if !reflect.DeepEqual(tables, original) {
		t.Fatalf("expected input to remain unchanged, got %v", tables)
	}
}
