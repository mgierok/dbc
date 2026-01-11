package usecase

import (
	"context"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/domain/model"
)

type SaveRecordEdits struct {
	engine port.Engine
}

func NewSaveRecordEdits(engine port.Engine) *SaveRecordEdits {
	return &SaveRecordEdits{engine: engine}
}

func (uc *SaveRecordEdits) Execute(ctx context.Context, tableName string, updates []model.RecordUpdate) error {
	return uc.engine.ApplyRecordUpdates(ctx, tableName, updates)
}
