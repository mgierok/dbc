package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
)

var (
	ErrConfigDatabaseNameRequired    = errors.New("config database name is required")
	ErrConfigDatabasePathRequired    = errors.New("config database path is required")
	ErrConfigDatabaseIndexOutOfRange = errors.New("config database index out of range")
)

type ListConfiguredDatabases struct {
	store port.ConfigStore
}

func NewListConfiguredDatabases(store port.ConfigStore) *ListConfiguredDatabases {
	return &ListConfiguredDatabases{store: store}
}

func (uc *ListConfiguredDatabases) Execute(ctx context.Context) ([]dto.ConfigDatabase, error) {
	entries, err := uc.store.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ConfigDatabase, len(entries))
	for i, entry := range entries {
		result[i] = dto.ConfigDatabase{
			Name: entry.Name,
			Path: entry.DBPath,
		}
	}
	return result, nil
}

type GetActiveConfigPath struct {
	store port.ConfigStore
}

func NewGetActiveConfigPath(store port.ConfigStore) *GetActiveConfigPath {
	return &GetActiveConfigPath{store: store}
}

func (uc *GetActiveConfigPath) Execute(ctx context.Context) (string, error) {
	return uc.store.ActivePath(ctx)
}

type CreateConfiguredDatabase struct {
	store port.ConfigStore
}

func NewCreateConfiguredDatabase(store port.ConfigStore) *CreateConfiguredDatabase {
	return &CreateConfiguredDatabase{store: store}
}

func (uc *CreateConfiguredDatabase) Execute(ctx context.Context, database dto.ConfigDatabase) error {
	entry, err := toConfigEntry(database)
	if err != nil {
		return err
	}
	return uc.store.Create(ctx, entry)
}

type UpdateConfiguredDatabase struct {
	store port.ConfigStore
}

func NewUpdateConfiguredDatabase(store port.ConfigStore) *UpdateConfiguredDatabase {
	return &UpdateConfiguredDatabase{store: store}
}

func (uc *UpdateConfiguredDatabase) Execute(ctx context.Context, index int, database dto.ConfigDatabase) error {
	if index < 0 {
		return ErrConfigDatabaseIndexOutOfRange
	}
	entry, err := toConfigEntry(database)
	if err != nil {
		return err
	}
	return uc.store.Update(ctx, index, entry)
}

type DeleteConfiguredDatabase struct {
	store port.ConfigStore
}

func NewDeleteConfiguredDatabase(store port.ConfigStore) *DeleteConfiguredDatabase {
	return &DeleteConfiguredDatabase{store: store}
}

func (uc *DeleteConfiguredDatabase) Execute(ctx context.Context, index int) error {
	if index < 0 {
		return ErrConfigDatabaseIndexOutOfRange
	}
	return uc.store.Delete(ctx, index)
}

func toConfigEntry(database dto.ConfigDatabase) (port.ConfigEntry, error) {
	name := strings.TrimSpace(database.Name)
	if name == "" {
		return port.ConfigEntry{}, ErrConfigDatabaseNameRequired
	}
	path := strings.TrimSpace(database.Path)
	if path == "" {
		return port.ConfigEntry{}, ErrConfigDatabasePathRequired
	}
	return port.ConfigEntry{Name: name, DBPath: path}, nil
}
