package usecase_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestLoadDatabaseSelectorState_ReturnsForcedSetupWhenNoOptionsExist(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{activePath: "/tmp/config.json"}
	uc := usecase.NewLoadDatabaseSelectorState(store)

	// Act
	state, err := uc.Execute(context.Background(), dto.DatabaseSelectorLoadInput{})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.ActiveConfigPath != "/tmp/config.json" {
		t.Fatalf("expected active config path %q, got %q", "/tmp/config.json", state.ActiveConfigPath)
	}
	if !state.RequiresFirstEntry {
		t.Fatal("expected forced setup when selector has no options")
	}
	if len(state.Options) != 0 {
		t.Fatalf("expected no selector options, got %d", len(state.Options))
	}
}

func TestLoadDatabaseSelectorState_MapsConfigEntriesInOrderWithConfigPermissions(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
			{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
		},
		activePath: "/tmp/config.json",
	}
	uc := usecase.NewLoadDatabaseSelectorState(store)

	// Act
	state, err := uc.Execute(context.Background(), dto.DatabaseSelectorLoadInput{})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := []dto.DatabaseSelectorOption{
		{
			Name:        "local",
			ConnString:  "/tmp/local.sqlite",
			Source:      dto.DatabaseSelectorOptionSourceConfig,
			ConfigIndex: 0,
			CanEdit:     true,
			CanDelete:   true,
		},
		{
			Name:        "analytics",
			ConnString:  "/tmp/analytics.sqlite",
			Source:      dto.DatabaseSelectorOptionSourceConfig,
			ConfigIndex: 1,
			CanEdit:     true,
			CanDelete:   true,
		},
	}
	if !reflect.DeepEqual(state.Options, expected) {
		t.Fatalf("expected selector options %+v, got %+v", expected, state.Options)
	}
	if state.RequiresFirstEntry {
		t.Fatal("expected forced setup to stay disabled when config entries exist")
	}
}

func TestLoadDatabaseSelectorState_NormalizesAndDeduplicatesAdditionalOptions(t *testing.T) {
	// Arrange
	basePath := filepath.Join(t.TempDir(), "direct.sqlite")
	uc := usecase.NewLoadDatabaseSelectorState(&fakeConfigStore{})

	// Act
	state, err := uc.Execute(context.Background(), dto.DatabaseSelectorLoadInput{
		AdditionalOptions: []dto.DatabaseSelectorAdditionalOption{
			{
				Name:       "   ",
				ConnString: " " + basePath + " ",
				Source:     dto.DatabaseSelectorOptionSourceConfig,
			},
			{
				Name:       "duplicate",
				ConnString: basePath + string(os.PathSeparator) + ".",
				Source:     dto.DatabaseSelectorOptionSourceCLI,
			},
			{
				Name:       "ignored",
				ConnString: "   ",
				Source:     dto.DatabaseSelectorOptionSourceCLI,
			},
		},
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := []dto.DatabaseSelectorOption{
		{
			Name:        basePath,
			ConnString:  basePath,
			Source:      dto.DatabaseSelectorOptionSourceCLI,
			ConfigIndex: -1,
			CanEdit:     false,
			CanDelete:   false,
		},
	}
	if !reflect.DeepEqual(state.Options, expected) {
		t.Fatalf("expected normalized selector options %+v, got %+v", expected, state.Options)
	}
}

func TestLoadDatabaseSelectorState_ConfigEntryWinsOverEquivalentAdditionalOption(t *testing.T) {
	// Arrange
	basePath := filepath.Join(t.TempDir(), "local.sqlite")
	store := &fakeConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: basePath},
		},
	}
	uc := usecase.NewLoadDatabaseSelectorState(store)

	// Act
	state, err := uc.Execute(context.Background(), dto.DatabaseSelectorLoadInput{
		AdditionalOptions: []dto.DatabaseSelectorAdditionalOption{
			{
				Name:       "cli duplicate",
				ConnString: basePath + string(os.PathSeparator) + ".",
				Source:     dto.DatabaseSelectorOptionSourceCLI,
			},
		},
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(state.Options) != 1 {
		t.Fatalf("expected one merged selector option, got %d", len(state.Options))
	}
	if state.Options[0].Name != "local" {
		t.Fatalf("expected config-backed name %q, got %q", "local", state.Options[0].Name)
	}
	if state.Options[0].Source != dto.DatabaseSelectorOptionSourceConfig {
		t.Fatalf("expected config-backed option, got %q", state.Options[0].Source)
	}
	if !state.Options[0].CanEdit || !state.Options[0].CanDelete {
		t.Fatalf("expected config-backed permissions to stay enabled, got %+v", state.Options[0])
	}
}
