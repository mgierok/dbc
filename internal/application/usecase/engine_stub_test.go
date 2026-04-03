package usecase_test

import (
	"context"

	"github.com/mgierok/dbc/internal/domain/model"
)

type engineStub struct {
	tables    []model.Table
	schema    model.Schema
	records   model.RecordPage
	operators []model.Operator

	listTablesErr    error
	getSchemaErr     error
	listRecordsErr   error
	listOperatorsErr error
	applyChangesErr  error

	lastRecordsTable  string
	lastRecordsOffset int
	lastRecordsLimit  int
	lastRecordsFilter *model.Filter
	lastRecordsSort   *model.Sort

	appliedTableName string
	appliedChanges   model.TableChanges
	appliedCount     int
}

func (s *engineStub) ListTables(context.Context) ([]model.Table, error) {
	if s.listTablesErr != nil {
		return nil, s.listTablesErr
	}
	return s.tables, nil
}

func (s *engineStub) GetSchema(_ context.Context, tableName string) (model.Schema, error) {
	if s.getSchemaErr != nil {
		return model.Schema{}, s.getSchemaErr
	}
	schema := s.schema
	schema.Table = model.Table{Name: tableName}
	return schema, nil
}

func (s *engineStub) ListRecords(_ context.Context, tableName string, offset, limit int, filter *model.Filter, sort *model.Sort) (model.RecordPage, error) {
	if s.listRecordsErr != nil {
		return model.RecordPage{}, s.listRecordsErr
	}

	s.lastRecordsTable = tableName
	s.lastRecordsOffset = offset
	s.lastRecordsLimit = limit
	s.lastRecordsFilter = filter
	s.lastRecordsSort = sort

	return s.records, nil
}

func (s *engineStub) ListOperators(context.Context, string) ([]model.Operator, error) {
	if s.listOperatorsErr != nil {
		return nil, s.listOperatorsErr
	}
	return s.operators, nil
}

func (s *engineStub) ApplyRecordChanges(_ context.Context, tableName string, changes model.TableChanges) (int, error) {
	s.appliedTableName = tableName
	s.appliedChanges = changes

	if s.applyChangesErr != nil {
		return 0, s.applyChangesErr
	}
	return s.appliedCount, nil
}
