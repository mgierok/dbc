package selector

import (
	"context"
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

func TestDatabaseSelector_AppendsSessionScopedAdditionalOptions(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
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

func TestDatabaseSelector_AdditionalOptionsSurviveRefresh(t *testing.T) {
	// Arrange
	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
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

	// Act
	if err := model.refreshOptions(); err != nil {
		t.Fatalf("expected refresh without error, got %v", err)
	}

	// Assert
	if len(model.options) != 2 {
		t.Fatalf("expected additional session option to survive refresh, got %d options", len(model.options))
	}
	if model.options[1].Source != DatabaseOptionSourceCLI {
		t.Fatalf("expected refreshed session option source %q, got %q", DatabaseOptionSourceCLI, model.options[1].Source)
	}
}
