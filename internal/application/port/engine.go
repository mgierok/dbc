package port

import (
	"context"

	"github.com/mgierok/dbc/internal/domain/model"
)

type Engine interface {
	ListTables(ctx context.Context) ([]model.Table, error)
	GetSchema(ctx context.Context, tableName string) (model.Schema, error)
	ListRecords(ctx context.Context, tableName string, offset, limit int, filter *model.Filter) (model.RecordPage, error)
	ListOperators(ctx context.Context, columnType string) ([]model.Operator, error)
	ApplyRecordUpdates(ctx context.Context, tableName string, updates []model.RecordUpdate) error
}
