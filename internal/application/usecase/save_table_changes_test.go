package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/domain/model"
)

func TestSaveTableChanges_DelegatesChangesAndReturnsAppliedRowCount(t *testing.T) {
	t.Parallel()

	engine := &engineStub{appliedCount: 1}
	uc := usecase.NewSaveTableChanges(engine)
	changes := model.TableChanges{
		Inserts: []model.RecordInsert{
			{
				Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "new", Raw: "new"}}},
			},
		},
	}

	count, err := uc.Execute(context.Background(), "users", changes)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("expected applied row count 1, got %d", count)
	}
	if engine.appliedTableName != "users" {
		t.Fatalf("expected table name %q, got %q", "users", engine.appliedTableName)
	}
	if len(engine.appliedChanges.Inserts) != 1 {
		t.Fatalf("expected 1 insert, got %d", len(engine.appliedChanges.Inserts))
	}
}

func TestSaveTableChanges_ReturnsError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	engine := &engineStub{applyChangesErr: expectedErr}
	uc := usecase.NewSaveTableChanges(engine)

	count, err := uc.Execute(context.Background(), "users", model.TableChanges{
		Updates: []model.RecordUpdate{
			{
				Identity: model.RecordIdentity{
					Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
				},
				Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
			},
		},
	})

	if count != 0 {
		t.Fatalf("expected zero count on error, got %d", count)
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSaveTableChanges_ValidateExecuteInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tableName string
		changes   model.TableChanges
		assertErr func(t *testing.T, err error)
	}{
		{
			name:      "missing table name",
			tableName: "   ",
			changes: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "alice", Raw: "alice"}}},
					},
				},
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatal("expected error for missing table name")
				}
				if err.Error() != "table name is required" {
					t.Fatalf("expected missing-table-name error, got %v", err)
				}
			},
		},
		{
			name:      "missing changes",
			tableName: "users",
			changes:   model.TableChanges{},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, model.ErrMissingTableChanges) {
					t.Fatalf("expected error %v, got %v", model.ErrMissingTableChanges, err)
				}
			},
		},
		{
			name:      "insert requires values",
			tableName: "users",
			changes: model.TableChanges{
				Inserts: []model.RecordInsert{{}},
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, model.ErrMissingInsertValues) {
					t.Fatalf("expected error %v, got %v", model.ErrMissingInsertValues, err)
				}
			},
		},
		{
			name:      "update requires identity",
			tableName: "users",
			changes: model.TableChanges{
				Updates: []model.RecordUpdate{
					{
						Changes: []model.ColumnValue{{Column: "name", Value: model.Value{Text: "bob", Raw: "bob"}}},
					},
				},
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, model.ErrMissingRecordIdentity) {
					t.Fatalf("expected error %v, got %v", model.ErrMissingRecordIdentity, err)
				}
			},
		},
		{
			name:      "update requires changes",
			tableName: "users",
			changes: model.TableChanges{
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}}},
						},
					},
				},
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, model.ErrMissingRecordChanges) {
					t.Fatalf("expected error %v, got %v", model.ErrMissingRecordChanges, err)
				}
			},
		},
		{
			name:      "delete requires identity",
			tableName: "users",
			changes: model.TableChanges{
				Deletes: []model.RecordDelete{{}},
			},
			assertErr: func(t *testing.T, err error) {
				t.Helper()
				if !errors.Is(err, model.ErrMissingDeleteIdentity) {
					t.Fatalf("expected error %v, got %v", model.ErrMissingDeleteIdentity, err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := usecase.NewSaveTableChanges(&engineStub{})

			count, err := uc.Execute(context.Background(), tc.tableName, tc.changes)

			if count != 0 {
				t.Fatalf("expected zero count for validation failure, got %d", count)
			}
			tc.assertErr(t, err)
		})
	}
}

func TestSaveTableChanges_ExecuteDTO_MapsPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		count           int
		changes         dto.TableChanges
		expectedChanges model.TableChanges
	}{
		{
			name:  "insert payload",
			count: 1,
			changes: dto.TableChanges{
				Inserts: []dto.RecordInsert{
					{
						Values: []dto.ColumnValue{{Column: "name", Value: dto.StagedValue{Text: "new", Raw: "new"}}},
					},
				},
			},
			expectedChanges: model.TableChanges{
				Inserts: []model.RecordInsert{
					{
						Values:             []model.ColumnValue{{Column: "name", Value: model.Value{Text: "new", Raw: "new"}}},
						ExplicitAutoValues: []model.ColumnValue{},
					},
				},
				Updates: []model.RecordUpdate{},
				Deletes: []model.RecordDelete{},
			},
		},
		{
			name: "update payload",
			changes: dto.TableChanges{
				Updates: []dto.RecordUpdate{
					{
						Identity: dto.RecordIdentity{
							Keys: []dto.RecordIdentityKey{
								{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
								{Column: "tenant_id", Value: dto.StagedValue{Text: "NULL", IsNull: true}},
							},
						},
						Changes: []dto.ColumnValue{
							{Column: "name", Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
						},
					},
				},
			},
			expectedChanges: model.TableChanges{
				Inserts: []model.RecordInsert{},
				Updates: []model.RecordUpdate{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{
								{Column: "id", Value: model.Value{Text: "1", Raw: int64(1)}},
								{Column: "tenant_id", Value: model.Value{Text: "NULL", IsNull: true}},
							},
						},
						Changes: []model.ColumnValue{
							{Column: "name", Value: model.Value{Text: "alice", Raw: "alice"}},
						},
					},
				},
				Deletes: []model.RecordDelete{},
			},
		},
		{
			name: "delete payload",
			changes: dto.TableChanges{
				Deletes: []dto.RecordDelete{
					{
						Identity: dto.RecordIdentity{
							Keys: []dto.RecordIdentityKey{
								{Column: "id", Value: dto.StagedValue{Text: "7", Raw: int64(7)}},
								{Column: "tenant_id", Value: dto.StagedValue{Text: "NULL", IsNull: true}},
							},
						},
					},
				},
			},
			expectedChanges: model.TableChanges{
				Inserts: []model.RecordInsert{},
				Updates: []model.RecordUpdate{},
				Deletes: []model.RecordDelete{
					{
						Identity: model.RecordIdentity{
							Keys: []model.RecordIdentityKey{
								{Column: "id", Value: model.Value{Text: "7", Raw: int64(7)}},
								{Column: "tenant_id", Value: model.Value{Text: "NULL", IsNull: true}},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			engine := &engineStub{appliedCount: tc.count}
			uc := usecase.NewSaveTableChanges(engine)

			count, err := uc.ExecuteDTO(context.Background(), "users", tc.changes)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if count != tc.count {
				t.Fatalf("expected applied row count %d, got %d", tc.count, count)
			}
			if engine.appliedTableName != "users" {
				t.Fatalf("expected table name %q, got %q", "users", engine.appliedTableName)
			}
			if !reflect.DeepEqual(engine.appliedChanges, tc.expectedChanges) {
				t.Fatalf("expected mapped changes %+v, got %+v", tc.expectedChanges, engine.appliedChanges)
			}
		})
	}
}

func TestSaveTableChanges_ExecuteDTO_PreservesCombinedPayloadShape(t *testing.T) {
	t.Parallel()

	engine := &engineStub{}
	uc := usecase.NewSaveTableChanges(engine)
	changes := dto.TableChanges{
		Updates: []dto.RecordUpdate{
			{
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
						{Column: "tenant_id", Value: dto.StagedValue{Text: "NULL", IsNull: true}},
					},
				},
				Changes: []dto.ColumnValue{
					{Column: "name", Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
				},
			},
		},
		Deletes: []dto.RecordDelete{
			{
				Identity: dto.RecordIdentity{
					Keys: []dto.RecordIdentityKey{
						{Column: "id", Value: dto.StagedValue{Text: "2", Raw: int64(2)}},
						{Column: "tenant_id", Value: dto.StagedValue{Text: "tenant-a", Raw: "tenant-a"}},
					},
				},
			},
		},
	}

	count, err := uc.ExecuteDTO(context.Background(), "users", changes)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected applied row count 0, got %d", count)
	}
	if engine.appliedTableName != "users" {
		t.Fatalf("expected table name users, got %q", engine.appliedTableName)
	}
	if len(engine.appliedChanges.Inserts) != 0 {
		t.Fatalf("expected 0 inserts, got %d", len(engine.appliedChanges.Inserts))
	}
	if len(engine.appliedChanges.Updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(engine.appliedChanges.Updates))
	}
	if len(engine.appliedChanges.Deletes) != 1 {
		t.Fatalf("expected 1 delete, got %d", len(engine.appliedChanges.Deletes))
	}
	if engine.appliedChanges.Updates[0].Identity.Keys[1].Value.Text != "NULL" {
		t.Fatalf("expected null update identity value to be preserved, got %+v", engine.appliedChanges.Updates[0].Identity.Keys[1].Value)
	}
	if engine.appliedChanges.Updates[0].Changes[0].Value.Raw != "alice" {
		t.Fatalf("expected update raw value alice, got %#v", engine.appliedChanges.Updates[0].Changes[0].Value.Raw)
	}
	if engine.appliedChanges.Deletes[0].Identity.Keys[1].Value.Raw != "tenant-a" {
		t.Fatalf("expected delete raw value tenant-a, got %#v", engine.appliedChanges.Deletes[0].Identity.Keys[1].Value.Raw)
	}
}
