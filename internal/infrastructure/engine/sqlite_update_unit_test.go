package engine

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/domain/model"
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

func TestPlanBatchChangeOperations_PreservesCallerOrderForRepeatedTableBatches(t *testing.T) {
	// Arrange
	batches := []namedTableChangeBatch{
		{
			tableName:     "users",
			incomingIndex: 0,
			changes: model.TableChanges{
				Deletes: []model.RecordDelete{
					{Identity: model.RecordIdentity{Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}}}},
				},
			},
		},
		{
			tableName:     "users",
			incomingIndex: 1,
			changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{Values: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}}},
				},
			},
		},
	}

	// Act
	operations := planBatchChangeOperations(batches, nil)

	// Assert
	expected := []string{
		"0:users:delete",
		"1:users:insert",
	}
	if got := plannedOperationLabels(operations); !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected planned operations %v, got %v", expected, got)
	}
}

func TestPlanBatchChangeOperations_ReordersForeignKeyDependentBatches(t *testing.T) {
	// Arrange
	batches := []namedTableChangeBatch{
		{
			tableName:     "books",
			incomingIndex: 0,
			changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{Values: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}}}},
				},
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}}},
						Changes:  []model.ColumnValue{{Column: "author_id", Value: model.Value{Text: "2", Raw: int64(2)}}},
					},
				},
			},
		},
		{
			tableName:     "authors",
			incomingIndex: 1,
			changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{Values: []model.ColumnValue{{Column: "id", Value: model.Value{Text: "2", Raw: int64(2)}}}},
				},
				Deletes: []model.RecordDelete{
					{Identity: model.RecordIdentity{Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}}}},
				},
			},
		},
	}
	parentsByChild := map[string]map[string]struct{}{
		"books": {"authors": {}},
	}

	// Act
	operations := planBatchChangeOperations(batches, parentsByChild)

	// Assert
	expected := []string{
		"1:authors:insert",
		"0:books:insert",
		"0:books:update",
		"1:authors:delete",
	}
	if got := plannedOperationLabels(operations); !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected planned operations %v, got %v", expected, got)
	}
}

func plannedOperationLabels(operations []plannedChangeOperation) []string {
	labels := make([]string, 0, len(operations))
	for _, operation := range operations {
		labels = append(labels, operationLabel(operation))
	}
	return labels
}

func operationLabel(operation plannedChangeOperation) string {
	return phaseLabel(operation.batch.incomingIndex, operation.batch.tableName, operation.phase)
}

func phaseLabel(batchIndex int, tableName string, phase changePhase) string {
	name := "unknown"
	switch phase {
	case changePhaseInsert:
		name = "insert"
	case changePhaseUpdate:
		name = "update"
	case changePhaseDelete:
		name = "delete"
	}
	return fmt.Sprintf("%d:%s:%s", batchIndex, tableName, name)
}
