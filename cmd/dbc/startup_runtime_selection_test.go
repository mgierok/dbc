package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func TestBuildConnectionFailureStatus_IncludesGuidanceAndDatabaseName(t *testing.T) {
	t.Parallel()

	// Arrange
	selected := tui.DatabaseOption{Name: "analytics"}

	// Act
	status := buildConnectionFailureStatus(selected, "ping failed")

	// Assert
	if !strings.Contains(status, "analytics") {
		t.Fatalf("expected selected database name in status, got %q", status)
	}
	if !strings.Contains(status, "Choose another database or edit selected entry") {
		t.Fatalf("expected user guidance in status, got %q", status)
	}
	if !strings.Contains(status, "ping failed") {
		t.Fatalf("expected error detail in status, got %q", status)
	}
}

func TestResolveStartupSelection_ChoosesExpectedStartupTarget(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name               string
		setup              func(t *testing.T) (startupOptions, []tui.DatabaseOption, tui.DatabaseOption, startupPath)
		wantListCalled     bool
		wantSelectorCalled bool
	}{
		{
			name: "uses direct launch without selector call",
			setup: func(t *testing.T) (startupOptions, []tui.DatabaseOption, tui.DatabaseOption, startupPath) {
				t.Helper()
				return startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}, nil, tui.DatabaseOption{
					Name:       "/tmp/direct.sqlite",
					ConnString: "/tmp/direct.sqlite",
					Source:     tui.DatabaseOptionSourceCLI,
				}, startupPathDirectLaunch
			},
			wantListCalled:     true,
			wantSelectorCalled: false,
		},
		{
			name: "uses selector when direct launch missing",
			setup: func(t *testing.T) (startupOptions, []tui.DatabaseOption, tui.DatabaseOption, startupPath) {
				t.Helper()
				return startupOptions{}, nil, tui.DatabaseOption{
					Name:       "analytics",
					ConnString: "/tmp/analytics.sqlite",
				}, startupPathSelector
			},
			wantListCalled:     false,
			wantSelectorCalled: true,
		},
		{
			name: "reuses configured identity when normalized paths match",
			setup: func(t *testing.T) (startupOptions, []tui.DatabaseOption, tui.DatabaseOption, startupPath) {
				t.Helper()
				configured := filepath.Join(t.TempDir(), "direct.sqlite")
				return startupOptions{directLaunchConnString: configured + string(os.PathSeparator) + "."}, []tui.DatabaseOption{
						{Name: "local", ConnString: configured, Source: tui.DatabaseOptionSourceConfig},
					}, tui.DatabaseOption{
						Name:       "local",
						ConnString: configured,
						Source:     tui.DatabaseOptionSourceConfig,
					}, startupPathDirectLaunch
			},
			wantListCalled:     true,
			wantSelectorCalled: false,
		},
		{
			name: "returns direct identity when no normalized configured match exists",
			setup: func(t *testing.T) (startupOptions, []tui.DatabaseOption, tui.DatabaseOption, startupPath) {
				t.Helper()
				return startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}, []tui.DatabaseOption{
						{Name: "local", ConnString: "/tmp/configured.sqlite", Source: tui.DatabaseOptionSourceConfig},
					}, tui.DatabaseOption{
						Name:       "/tmp/direct.sqlite",
						ConnString: "/tmp/direct.sqlite",
						Source:     tui.DatabaseOptionSourceCLI,
					}, startupPathDirectLaunch
			},
			wantListCalled:     true,
			wantSelectorCalled: false,
		},
		{
			name: "uses first configured identity for deterministic normalized match",
			setup: func(t *testing.T) (startupOptions, []tui.DatabaseOption, tui.DatabaseOption, startupPath) {
				t.Helper()
				configured := filepath.Join(t.TempDir(), "direct.sqlite")
				return startupOptions{directLaunchConnString: configured + string(os.PathSeparator) + "."}, []tui.DatabaseOption{
						{Name: "first", ConnString: configured, Source: tui.DatabaseOptionSourceConfig},
						{Name: "second", ConnString: configured + string(os.PathSeparator) + ".", Source: tui.DatabaseOptionSourceConfig},
					}, tui.DatabaseOption{
						Name:       "first",
						ConnString: configured,
						Source:     tui.DatabaseOptionSourceConfig,
					}, startupPathDirectLaunch
			},
			wantListCalled:     true,
			wantSelectorCalled: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			options, configuredOptions, selectorResult, wantPath := tc.setup(t)
			listCalled := false
			selectorCalled := false

			// Act
			selected, path, err := resolveStartupSelection(
				options,
				func() ([]tui.DatabaseOption, error) {
					listCalled = true
					return configuredOptions, nil
				},
				func() (tui.DatabaseOption, error) {
					selectorCalled = true
					return selectorResult, nil
				},
			)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if listCalled != tc.wantListCalled {
				t.Fatalf("expected list callback called=%t, got %t", tc.wantListCalled, listCalled)
			}
			if selectorCalled != tc.wantSelectorCalled {
				t.Fatalf("expected selector callback called=%t, got %t", tc.wantSelectorCalled, selectorCalled)
			}
			if path != wantPath {
				t.Fatalf("expected startup path %v, got %v", wantPath, path)
			}
			assertDatabaseOption(t, selected, selectorResult)
		})
	}
}

func TestResolveStartupSelection_PropagatesSelectionErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		options   startupOptions
		listErr   error
		selectErr error
		wantPath  startupPath
	}{
		{
			name:     "returns direct launch list error",
			options:  startupOptions{directLaunchConnString: "/tmp/direct.sqlite"},
			listErr:  errors.New("list failed"),
			wantPath: startupPathDirectLaunch,
		},
		{
			name:      "returns selector error",
			options:   startupOptions{},
			selectErr: errors.New("selector failed"),
			wantPath:  startupPathSelector,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			_, path, err := resolveStartupSelection(
				tc.options,
				func() ([]tui.DatabaseOption, error) {
					if tc.listErr != nil {
						return nil, tc.listErr
					}
					return nil, nil
				},
				func() (tui.DatabaseOption, error) {
					return tui.DatabaseOption{}, tc.selectErr
				},
			)

			// Assert
			if path != tc.wantPath {
				t.Fatalf("expected startup path %v, got %v", tc.wantPath, path)
			}
			switch {
			case tc.listErr != nil && !errors.Is(err, tc.listErr):
				t.Fatalf("expected list error %v, got %v", tc.listErr, err)
			case tc.selectErr != nil && !errors.Is(err, tc.selectErr):
				t.Fatalf("expected selector error %v, got %v", tc.selectErr, err)
			}
		})
	}
}

func TestNewStartupSelectionStrategy_ResolvesExpectedStartupPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		directLaunchPending bool
		wantListCalled      bool
		wantSelectorCalled  bool
		wantPath            startupPath
		wantOption          tui.DatabaseOption
	}{
		{
			name:                "uses direct launch when pending",
			directLaunchPending: true,
			wantListCalled:      true,
			wantSelectorCalled:  false,
			wantPath:            startupPathDirectLaunch,
			wantOption: tui.DatabaseOption{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     tui.DatabaseOptionSourceCLI,
			},
		},
		{
			name:                "falls back to selector when direct launch is not pending",
			directLaunchPending: false,
			wantListCalled:      false,
			wantSelectorCalled:  true,
			wantPath:            startupPathSelector,
			wantOption: tui.DatabaseOption{
				Name:       "analytics",
				ConnString: "/tmp/analytics.sqlite",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			strategy := newStartupSelectionStrategy(startupOptions{directLaunchConnString: "/tmp/direct.sqlite"}, tc.directLaunchPending)
			listCalled := false
			selectorCalled := false

			// Act
			selected, path, err := strategy.resolve(
				func() ([]tui.DatabaseOption, error) {
					listCalled = true
					return nil, nil
				},
				func() (tui.DatabaseOption, error) {
					selectorCalled = true
					return tc.wantOption, nil
				},
			)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if listCalled != tc.wantListCalled {
				t.Fatalf("expected list callback called=%t, got %t", tc.wantListCalled, listCalled)
			}
			if selectorCalled != tc.wantSelectorCalled {
				t.Fatalf("expected selector callback called=%t, got %t", tc.wantSelectorCalled, selectorCalled)
			}
			if path != tc.wantPath {
				t.Fatalf("expected startup path %v, got %v", tc.wantPath, path)
			}
			assertDatabaseOption(t, selected, tc.wantOption)
		})
	}
}

func TestListConfiguredDatabaseOptions_MapsConfigEntriesToSelectorOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	store := &fakeStartupSelectionConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
			{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
		},
	}
	listConfiguredDatabases := usecase.NewListConfiguredDatabases(store)

	// Act
	options, err := listConfiguredDatabaseOptions(context.Background(), listConfiguredDatabases)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(options) != 2 {
		t.Fatalf("expected two mapped options, got %d", len(options))
	}
	if options[0].Name != "local" || options[0].ConnString != "/tmp/local.sqlite" || options[0].Source != tui.DatabaseOptionSourceConfig {
		t.Fatalf("unexpected first option mapping: %+v", options[0])
	}
	if options[1].Name != "analytics" || options[1].ConnString != "/tmp/analytics.sqlite" || options[1].Source != tui.DatabaseOptionSourceConfig {
		t.Fatalf("unexpected second option mapping: %+v", options[1])
	}
}

func TestBuildDirectLaunchFailureMessage_ContainsReasonAndGuidance(t *testing.T) {
	t.Parallel()

	// Act
	message := buildDirectLaunchFailureMessage("/tmp/missing.sqlite", "database file does not exist")

	// Assert
	if !strings.Contains(message, "database file does not exist") {
		t.Fatalf("expected failure reason in message, got %q", message)
	}
	if !strings.Contains(message, "retry with -d/--database") {
		t.Fatalf("expected corrective guidance in message, got %q", message)
	}
}

func TestTrackSessionScopedCLIOption_TracksOnlyUniqueCLIOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		setup    func(t *testing.T) ([]tui.DatabaseOption, tui.DatabaseOption)
		wantSize int
	}{
		{
			name: "adds direct cli selection",
			setup: func(t *testing.T) ([]tui.DatabaseOption, tui.DatabaseOption) {
				t.Helper()
				return nil, tui.DatabaseOption{
					Name:       "/tmp/direct.sqlite",
					ConnString: "/tmp/direct.sqlite",
					Source:     tui.DatabaseOptionSourceCLI,
				}
			},
			wantSize: 1,
		},
		{
			name: "skips configured selection",
			setup: func(t *testing.T) ([]tui.DatabaseOption, tui.DatabaseOption) {
				t.Helper()
				return nil, tui.DatabaseOption{
					Name:       "local",
					ConnString: "/tmp/local.sqlite",
					Source:     tui.DatabaseOptionSourceConfig,
				}
			},
			wantSize: 0,
		},
		{
			name: "deduplicates normalized cli identity",
			setup: func(t *testing.T) ([]tui.DatabaseOption, tui.DatabaseOption) {
				t.Helper()
				basePath := filepath.Join(t.TempDir(), "direct.sqlite")
				return []tui.DatabaseOption{
						{
							Name:       basePath,
							ConnString: basePath,
							Source:     tui.DatabaseOptionSourceCLI,
						},
					}, tui.DatabaseOption{
						Name:       basePath + string(os.PathSeparator) + ".",
						ConnString: basePath + string(os.PathSeparator) + ".",
						Source:     tui.DatabaseOptionSourceCLI,
					}
			},
			wantSize: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			existing, selected := tc.setup(t)

			// Act
			updated := trackSessionScopedCLIOption(existing, selected)

			// Assert
			if len(updated) != tc.wantSize {
				t.Fatalf("expected %d session-scoped options, got %d", tc.wantSize, len(updated))
			}
		})
	}
}

func TestTrackSessionScopedDirectLaunchOption_DelegatesToCLITracking(t *testing.T) {
	t.Parallel()

	// Arrange
	selected := tui.DatabaseOption{
		Name:       "/tmp/direct.sqlite",
		ConnString: "/tmp/direct.sqlite",
		Source:     tui.DatabaseOptionSourceCLI,
	}

	// Act
	updated := trackSessionScopedDirectLaunchOption(nil, startupPathDirectLaunch, selected)

	// Assert
	if len(updated) != 1 {
		t.Fatalf("expected one session-scoped option, got %d", len(updated))
	}
	assertDatabaseOption(t, updated[0], selected)
}

type fakeStartupSelectionConfigStore struct {
	entries []port.ConfigEntry
}

func (f *fakeStartupSelectionConfigStore) List(_ context.Context) ([]port.ConfigEntry, error) {
	result := make([]port.ConfigEntry, len(f.entries))
	copy(result, f.entries)
	return result, nil
}

func (f *fakeStartupSelectionConfigStore) Create(_ context.Context, entry port.ConfigEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeStartupSelectionConfigStore) Update(_ context.Context, index int, entry port.ConfigEntry) error {
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.entries[index] = entry
	return nil
}

func (f *fakeStartupSelectionConfigStore) Delete(_ context.Context, index int) error {
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.entries = append(f.entries[:index], f.entries[index+1:]...)
	return nil
}

func (f *fakeStartupSelectionConfigStore) ActivePath(_ context.Context) (string, error) {
	return "/tmp/config.json", nil
}
