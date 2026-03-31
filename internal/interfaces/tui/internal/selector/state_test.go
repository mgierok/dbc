package selector

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestDatabaseSelector_EmptyConfigStartsInForcedSetupForm(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager)

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if !model.requiresFirstEntry {
		t.Fatal("expected forced setup mode when no configured databases exist")
	}
	if model.mode != selectorModeAdd {
		t.Fatalf("expected add form mode in forced setup, got %v", model.mode)
	}
}

func TestDatabaseSelector_ForcedSetupUsesLoadStateInsteadOfAdapterLocalMerge(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "ignored", Path: "/tmp/ignored.sqlite"},
		},
		loadState: &dto.DatabaseSelectorState{
			ActiveConfigPath:   "/tmp/config.json",
			RequiresFirstEntry: true,
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager)

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if !model.requiresFirstEntry {
		t.Fatal("expected forced setup flag from loaded state")
	}
	if model.mode != selectorModeAdd {
		t.Fatalf("expected add form mode in forced setup, got %v", model.mode)
	}
}

func TestDatabaseSelector_AppliesInitialStatusMessage(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		StatusMessage: "Connection failed: invalid path",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if model.browse.statusMessage != "Connection failed: invalid path" {
		t.Fatalf("expected initial status message, got %q", model.browse.statusMessage)
	}
}

func TestDatabaseSelector_PrefersSelectionByConnString(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
			{Name: "analytics", Path: "/tmp/analytics.sqlite"},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		PreferConnString: "/tmp/analytics.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if model.browse.selected != 1 {
		t.Fatalf("expected preferred selection index %d, got %d", 1, model.browse.selected)
	}
}

func TestDatabaseSelector_PrefersSelectionByEquivalentConnString(t *testing.T) {
	// Arrange
	basePath := filepath.Join(t.TempDir(), "analytics.sqlite")
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
			{Name: "analytics", Path: basePath},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		PreferConnString: basePath + string(os.PathSeparator) + ".",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if model.browse.selected != 1 {
		t.Fatalf("expected preferred equivalent selection index %d, got %d", 1, model.browse.selected)
	}
}

func TestDatabaseSelector_IgnoresMissingPreferredConnString(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
			{Name: "analytics", Path: "/tmp/analytics.sqlite"},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		PreferConnString: "/tmp/missing.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if model.browse.selected != 0 {
		t.Fatalf("expected default selection index %d, got %d", 0, model.browse.selected)
	}
}

func TestDatabaseSelector_PassesAdditionalOptionsToLoadState(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		loadState: &dto.DatabaseSelectorState{
			ActiveConfigPath: "/tmp/config.json",
			Options: []dto.DatabaseSelectorOption{
				{
					Name:        "local",
					ConnString:  "/tmp/local.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceConfig,
					ConfigIndex: 0,
					CanEdit:     true,
					CanDelete:   true,
				},
			},
		},
	}
	expectedInput := dto.DatabaseSelectorLoadInput{
		AdditionalOptions: []dto.DatabaseSelectorAdditionalOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     dto.DatabaseSelectorOptionSourceCLI,
			},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if model == nil {
		t.Fatal("expected selector model instance")
	}
	if !reflect.DeepEqual(manager.lastLoadInput, expectedInput) {
		t.Fatalf("expected selector load input %+v, got %+v", expectedInput, manager.lastLoadInput)
	}
}

func TestDatabaseSelector_UsesLoadedStateToRenderSessionScopedAdditionalOptions(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		loadState: &dto.DatabaseSelectorState{
			ActiveConfigPath: "/tmp/config.json",
			Options: []dto.DatabaseSelectorOption{
				{
					Name:        "local",
					ConnString:  "/tmp/local.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceConfig,
					ConfigIndex: 0,
					CanEdit:     true,
					CanDelete:   true,
				},
				{
					Name:        "/tmp/direct.sqlite",
					ConnString:  "/tmp/direct.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceCLI,
					ConfigIndex: -1,
					CanEdit:     false,
					CanDelete:   false,
				},
			},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if len(model.options) != 2 {
		t.Fatalf("expected two options with one session-scoped entry, got %d", len(model.options))
	}
	if model.options[0].Source != DatabaseOptionSourceConfig {
		t.Fatalf("expected first option to be config-backed, got %q", model.options[0].Source)
	}
	if model.options[1].Source != DatabaseOptionSourceCLI {
		t.Fatalf("expected second option to be CLI session entry, got %q", model.options[1].Source)
	}
}

func TestDatabaseSelector_RefreshOptionsReloadsStateUsingLaunchAdditionalOptions(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		loadState: &dto.DatabaseSelectorState{
			ActiveConfigPath: "/tmp/config.json",
			Options: []dto.DatabaseSelectorOption{
				{
					Name:        "local",
					ConnString:  "/tmp/local.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceConfig,
					ConfigIndex: 0,
					CanEdit:     true,
					CanDelete:   true,
				},
				{
					Name:        "/tmp/direct.sqlite",
					ConnString:  "/tmp/direct.sqlite",
					Source:      dto.DatabaseSelectorOptionSourceCLI,
					ConfigIndex: -1,
					CanEdit:     false,
					CanDelete:   false,
				},
			},
		},
	}
	expectedInput := dto.DatabaseSelectorLoadInput{
		AdditionalOptions: []dto.DatabaseSelectorAdditionalOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     dto.DatabaseSelectorOptionSourceCLI,
			},
		},
	}
	model, err := newDatabaseSelectorModel(context.Background(), manager, SelectorLaunchState{
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	manager.loadState = &dto.DatabaseSelectorState{
		ActiveConfigPath: "/tmp/config.json",
		Options: []dto.DatabaseSelectorOption{
			{
				Name:        "analytics",
				ConnString:  "/tmp/analytics.sqlite",
				Source:      dto.DatabaseSelectorOptionSourceConfig,
				ConfigIndex: 0,
				CanEdit:     true,
				CanDelete:   true,
			},
			{
				Name:        "/tmp/direct.sqlite",
				ConnString:  "/tmp/direct.sqlite",
				Source:      dto.DatabaseSelectorOptionSourceCLI,
				ConfigIndex: -1,
				CanEdit:     false,
				CanDelete:   false,
			},
		},
	}

	// Act
	if err := model.refreshOptions(); err != nil {
		t.Fatalf("expected refresh without error, got %v", err)
	}

	// Assert
	if !reflect.DeepEqual(manager.lastLoadInput, expectedInput) {
		t.Fatalf("expected refreshed selector load input %+v, got %+v", expectedInput, manager.lastLoadInput)
	}
	if len(model.options) != 2 {
		t.Fatalf("expected refreshed selector options, got %d options", len(model.options))
	}
	if model.options[0].Name != "analytics" {
		t.Fatalf("expected refreshed config option name %q, got %q", "analytics", model.options[0].Name)
	}
	if model.options[1].Source != DatabaseOptionSourceCLI {
		t.Fatalf("expected refreshed session option source %q, got %q", DatabaseOptionSourceCLI, model.options[1].Source)
	}
}

func TestDatabaseSelector_ReturnsErrorWhenListingEntriesFails(t *testing.T) {
	// Arrange
	loadStateErr := errors.New("load state failed")
	manager := &fakeSelectorManager{
		loadStateErr: loadStateErr,
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager)

	// Assert
	if model != nil {
		t.Fatalf("expected nil model when list fails, got %#v", model)
	}
	if !errors.Is(err, loadStateErr) {
		t.Fatalf("expected load-state error %v, got %v", loadStateErr, err)
	}
}

func TestDatabaseSelector_ReturnsErrorWhenLoadingActiveConfigPathFails(t *testing.T) {
	// Arrange
	activePathErr := errors.New("active path failed")
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
		activePathErr: activePathErr,
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager)

	// Assert
	if model != nil {
		t.Fatalf("expected nil model when active config path lookup fails, got %#v", model)
	}
	if !errors.Is(err, activePathErr) {
		t.Fatalf("expected active path error %v, got %v", activePathErr, err)
	}
}
