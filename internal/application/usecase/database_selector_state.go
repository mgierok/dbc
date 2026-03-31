package usecase

import (
	"context"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/sqliteidentity"
)

type LoadDatabaseSelectorState struct {
	store port.ConfigStore
}

func NewLoadDatabaseSelectorState(store port.ConfigStore) *LoadDatabaseSelectorState {
	return &LoadDatabaseSelectorState{store: store}
}

func (uc *LoadDatabaseSelectorState) Execute(ctx context.Context, input dto.DatabaseSelectorLoadInput) (dto.DatabaseSelectorState, error) {
	entries, err := uc.store.List(ctx)
	if err != nil {
		return dto.DatabaseSelectorState{}, err
	}

	activeConfigPath, err := uc.store.ActivePath(ctx)
	if err != nil {
		return dto.DatabaseSelectorState{}, err
	}

	configOptions := make([]dto.DatabaseSelectorOption, len(entries))
	seen := make(map[string]struct{}, len(entries)+len(input.AdditionalOptions))
	for i, entry := range entries {
		configOptions[i] = dto.DatabaseSelectorOption{
			Name:        entry.Name,
			ConnString:  entry.DBPath,
			Source:      dto.DatabaseSelectorOptionSourceConfig,
			ConfigIndex: i,
			CanEdit:     true,
			CanDelete:   true,
		}
		identity := sqliteidentity.Normalize(entry.DBPath)
		if identity == "" {
			continue
		}
		seen[identity] = struct{}{}
	}

	additionalOptions := normalizeSelectorAdditionalOptions(input.AdditionalOptions)
	options := make([]dto.DatabaseSelectorOption, 0, len(configOptions)+len(additionalOptions))
	options = append(options, configOptions...)
	for _, option := range additionalOptions {
		identity := sqliteidentity.Normalize(option.ConnString)
		if identity != "" {
			if _, exists := seen[identity]; exists {
				continue
			}
			seen[identity] = struct{}{}
		}
		options = append(options, option)
	}

	return dto.DatabaseSelectorState{
		ActiveConfigPath:   activeConfigPath,
		Options:            options,
		RequiresFirstEntry: len(options) == 0,
	}, nil
}

func normalizeSelectorAdditionalOptions(options []dto.DatabaseSelectorAdditionalOption) []dto.DatabaseSelectorOption {
	if len(options) == 0 {
		return nil
	}

	normalized := make([]dto.DatabaseSelectorOption, 0, len(options))
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		connString := strings.TrimSpace(option.ConnString)
		identity := sqliteidentity.Normalize(connString)
		if identity == "" {
			continue
		}
		if _, exists := seen[identity]; exists {
			continue
		}
		seen[identity] = struct{}{}

		name := strings.TrimSpace(option.Name)
		if name == "" {
			name = connString
		}
		normalized = append(normalized, dto.DatabaseSelectorOption{
			Name:        name,
			ConnString:  connString,
			Source:      dto.DatabaseSelectorOptionSourceCLI,
			ConfigIndex: -1,
			CanEdit:     false,
			CanDelete:   false,
		})
	}
	return normalized
}
