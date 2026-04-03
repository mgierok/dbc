package tui

import (
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func newRecordsViewModel(schema dto.Schema, records []dto.RecordRow) *Model {
	return &Model{
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema:   schema,
			records:  records,
		},
	}
}

func newStyledRecordsViewModel(schema dto.Schema, records []dto.RecordRow) *Model {
	model := newRecordsViewModel(schema, records)
	model.styles = primitives.NewRenderStyles(true)
	return model
}

func newRecordDetailModel(schema dto.Schema, records []dto.RecordRow) *Model {
	model := newStyledRecordsViewModel(schema, records)
	model.overlay.recordDetail.active = true
	return model
}

func mustPersistedRecordKey(t *testing.T, model *Model, rowIndex int) string {
	t.Helper()

	key, ok := model.recordKeyForPersistedRow(rowIndex)
	if !ok {
		t.Fatalf("expected persisted row key for row %d", rowIndex)
	}

	return key
}
