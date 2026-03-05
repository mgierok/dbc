package usecase

import "github.com/mgierok/dbc/internal/application/dto"

type StagingPolicy struct{}

func NewStagingPolicy() *StagingPolicy {
	return &StagingPolicy{}
}

func (p *StagingPolicy) InitialInsertValue(column dto.SchemaColumn) dto.StagedValue {
	if column.DefaultValue != nil {
		return dto.StagedValue{Text: *column.DefaultValue, Raw: *column.DefaultValue}
	}
	if column.Nullable {
		return dto.StagedValue{IsNull: true, Text: "NULL"}
	}
	return dto.StagedValue{Text: "", Raw: ""}
}

func (p *StagingPolicy) DirtyEditCount(
	pendingInserts []dto.PendingInsertRow,
	pendingUpdates map[string]dto.PendingRecordEdits,
	pendingDeletes map[string]dto.PendingRecordDelete,
) int {
	count := 0
	for _, edits := range pendingUpdates {
		count += len(edits.Changes)
	}
	count += len(pendingInserts)
	count += len(pendingDeletes)
	return count
}
