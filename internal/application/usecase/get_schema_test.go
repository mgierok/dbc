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

func TestGetSchema_MapsColumns(t *testing.T) {
	t.Parallel()

	defaultName := "'guest'"
	engine := &engineStub{
		schema: model.Schema{
			Columns: []model.Column{
				{
					Name:          "id",
					Type:          "INTEGER",
					Nullable:      false,
					PrimaryKey:    true,
					Unique:        true,
					AutoIncrement: true,
					ForeignKeys: []model.ForeignKeyRef{
						{Table: "accounts", Column: "owner_id"},
					},
				},
				{Name: "name", Type: "TEXT", Nullable: true},
				{
					Name:         "display_name",
					Type:         "TEXT",
					Nullable:     false,
					DefaultValue: &defaultName,
					Unique:       true,
				},
			},
		},
	}
	uc := usecase.NewGetSchema(engine)

	result, err := uc.Execute(context.Background(), "users")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.TableName != "users" {
		t.Fatalf("expected table name %q, got %q", "users", result.TableName)
	}

	expectedColumns := []dto.SchemaColumn{
		{
			Name:          "id",
			Type:          "INTEGER",
			Nullable:      false,
			PrimaryKey:    true,
			Unique:        true,
			AutoIncrement: true,
			ForeignKeys: []dto.ForeignKeyRef{
				{Table: "accounts", Column: "owner_id"},
			},
			MetadataBadges: []string{"PK", "NOT NULL", "AUTOINCREMENT", "FK->accounts.owner_id"},
			Input:          dto.ColumnInput{Kind: dto.ColumnInputText},
		},
		{
			Name:           "name",
			Type:           "TEXT",
			Nullable:       true,
			AutoIncrement:  false,
			MetadataBadges: []string{"NULL"},
			Input:          dto.ColumnInput{Kind: dto.ColumnInputText},
		},
		{
			Name:           "display_name",
			Type:           "TEXT",
			Nullable:       false,
			DefaultValue:   &defaultName,
			AutoIncrement:  false,
			Unique:         true,
			MetadataBadges: []string{"NOT NULL", "UNIQUE", "DEFAULT 'guest'"},
			Input:          dto.ColumnInput{Kind: dto.ColumnInputText},
		},
	}
	if !reflect.DeepEqual(result.Columns, expectedColumns) {
		t.Fatalf("expected %v, got %v", expectedColumns, result.Columns)
	}
}

func TestGetSchema_MapsForeignKeyBadgeWithoutReferencedColumn(t *testing.T) {
	t.Parallel()

	engine := &engineStub{
		schema: model.Schema{
			Columns: []model.Column{
				{
					Name:     "account_id",
					Type:     "INTEGER",
					Nullable: false,
					ForeignKeys: []model.ForeignKeyRef{
						{Table: "accounts"},
					},
				},
			},
		},
	}
	uc := usecase.NewGetSchema(engine)

	result, err := uc.Execute(context.Background(), "users")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"NOT NULL", "FK->accounts"}
	if !reflect.DeepEqual(result.Columns[0].MetadataBadges, expected) {
		t.Fatalf("expected metadata badges %v, got %v", expected, result.Columns[0].MetadataBadges)
	}
}

func TestGetSchema_PropagatesEngineError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("get schema failed")
	uc := usecase.NewGetSchema(&engineStub{getSchemaErr: expectedErr})

	_, err := uc.Execute(context.Background(), "users")

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
