package tui

import (
	"context"

	"github.com/mgierok/dbc/internal/application/dto"
)

type spyListRecordsUseCase struct {
	lastSort          *dto.Sort
	lastFilter        *dto.Filter
	lastRecordsOffset int
	lastRecordsLimit  int
	page              dto.RecordPage
	err               error
}

func (s *spyListRecordsUseCase) Execute(ctx context.Context, tableName string, offset, limit int, filter *dto.Filter, sort *dto.Sort) (dto.RecordPage, error) {
	s.lastSort = sort
	if filter != nil {
		copied := *filter
		s.lastFilter = &copied
	} else {
		s.lastFilter = nil
	}
	s.lastRecordsOffset = offset
	s.lastRecordsLimit = limit
	if s.err != nil {
		return dto.RecordPage{}, s.err
	}
	return s.page, nil
}

type spyListOperatorsUseCase struct {
	operators      []dto.Operator
	err            error
	lastColumnType string
}

func (s *spyListOperatorsUseCase) Execute(ctx context.Context, columnType string) ([]dto.Operator, error) {
	s.lastColumnType = columnType
	if s.err != nil {
		return nil, s.err
	}
	return append([]dto.Operator(nil), s.operators...), nil
}

type spySaveChangesUseCase struct {
	lastChanges []dto.NamedTableChanges
	err         error
}

func (s *spySaveChangesUseCase) ExecuteDTO(ctx context.Context, changes []dto.NamedTableChanges) error {
	s.lastChanges = append([]dto.NamedTableChanges(nil), changes...)
	return s.err
}

func testDatabaseStaging(states map[string]stagingState) databaseStagingState {
	converted := make(map[string]*stagingState, len(states))
	for tableName, state := range states {
		cloned := state
		converted[tableName] = &cloned
	}
	return databaseStagingState{tables: converted}
}

func testActiveDatabaseStaging(state stagingState) databaseStagingState {
	return testDatabaseStaging(map[string]stagingState{"": state})
}
