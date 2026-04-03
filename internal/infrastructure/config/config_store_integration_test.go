package config_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestStore_ListReturnsEntriesFromConfig(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"},{"name":"analytics","db_path":"/tmp/analytics.sqlite"}]}`)
	store := config.NewStore(path)

	// Act
	entries, err := store.List(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertConfigEntries(t, entries, []port.ConfigEntry{
		{Name: "local", DBPath: "/tmp/local.sqlite"},
		{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
	})
}

func TestStore_ListReturnsEmptyForStartupCompatibleEmptyStates(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, path string)
	}{
		{
			name: "missing file",
			setup: func(*testing.T, string) {
			},
		},
		{
			name: "trimmed empty file",
			setup: func(t *testing.T, path string) {
				writeConfigFile(t, path, " \n\t ")
			},
		},
		{
			name: "empty databases list",
			setup: func(t *testing.T, path string) {
				writeConfigFile(t, path, `{"databases":[]}`)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			path := filepath.Join(t.TempDir(), "config.json")
			tc.setup(t, path)
			store := config.NewStore(path)

			// Act
			entries, err := store.List(context.Background())

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			assertConfigEntries(t, entries, nil)
		})
	}
}

func TestStore_ListReturnsErrorWhenConfigUsesUnknownShape(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"database":{"name":"legacy","db_path":"/tmp/legacy.sqlite"}}`)
	store := config.NewStore(path)

	// Act
	_, err := store.List(context.Background())

	// Assert
	if err == nil {
		t.Fatal("expected malformed config error, got nil")
	}
}

func TestStore_CreatePersistsNewEntry(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`)
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   "analytics",
		DBPath: "/tmp/analytics.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertStoredEntries(t, path, []port.ConfigEntry{
		{Name: "local", DBPath: "/tmp/local.sqlite"},
		{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
	})
}

func TestStore_CreateInitializesStartupCompatibleEmptyStates(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, path string)
	}{
		{
			name: "missing file",
			setup: func(*testing.T, string) {
			},
		},
		{
			name: "trimmed empty file",
			setup: func(t *testing.T, path string) {
				writeConfigFile(t, path, " \n\t ")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			path := filepath.Join(t.TempDir(), "config.json")
			tc.setup(t, path)
			store := config.NewStore(path)

			// Act
			err := store.Create(context.Background(), port.ConfigEntry{
				Name:   "local",
				DBPath: "/tmp/local.sqlite",
			})

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			assertStoredEntries(t, path, []port.ConfigEntry{
				{Name: "local", DBPath: "/tmp/local.sqlite"},
			})
		})
	}
}

func TestStore_CreateCreatesConfigDirectory(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "nested", "dbc", "config.json")
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   "local",
		DBPath: "/tmp/local.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	info, statErr := os.Stat(filepath.Dir(path))
	if statErr != nil {
		t.Fatalf("expected config directory to exist, got %v", statErr)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", filepath.Dir(path))
	}
}

func TestStore_CreateRejectsInvalidEntry(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`)
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   " ",
		DBPath: "/tmp/analytics.sqlite",
	})

	// Assert
	if !errors.Is(err, config.ErrMissingDatabaseName) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabaseName, err)
	}
	assertStoredEntries(t, path, []port.ConfigEntry{
		{Name: "local", DBPath: "/tmp/local.sqlite"},
	})
}

func TestStore_CreateReturnsErrorWhenConfigHasInvalidSyntax(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[`)
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   "local",
		DBPath: "/tmp/local.sqlite",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid config syntax, got nil")
	}
}

func TestStore_CreateEnforcesSerializedConfigSizeLimit(t *testing.T) {
	testCases := []struct {
		name    string
		entry   port.ConfigEntry
		wantErr error
	}{
		{
			name:  "at limit",
			entry: configEntryForSerializedSize(t, config.Config{}, 1<<20, true),
		},
		{
			name:    "above limit",
			entry:   configEntryForSerializedSize(t, config.Config{}, 1<<20, false),
			wantErr: config.ErrConfigTooLarge,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			path := filepath.Join(t.TempDir(), "config.json")
			store := config.NewStore(path)

			// Act
			err := store.Create(context.Background(), tc.entry)

			// Assert
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("expected config file to remain absent, got stat error %v", statErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			info, statErr := os.Stat(path)
			if statErr != nil {
				t.Fatalf("failed to stat written config file: %v", statErr)
			}
			if info.Size() != 1<<20 {
				t.Fatalf("expected written file size %d, got %d", 1<<20, info.Size())
			}
			assertStoredEntries(t, path, []port.ConfigEntry{tc.entry})
		})
	}
}

func TestStore_UpdatePersistsReplacementEntry(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"},{"name":"analytics","db_path":"/tmp/analytics.sqlite"}]}`)
	store := config.NewStore(path)

	// Act
	err := store.Update(context.Background(), 0, port.ConfigEntry{
		Name:   "primary",
		DBPath: "/tmp/primary.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertStoredEntries(t, path, []port.ConfigEntry{
		{Name: "primary", DBPath: "/tmp/primary.sqlite"},
		{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
	})
}

func TestStore_UpdateReturnsErrorForIndexOutOfRange(t *testing.T) {
	testCases := []struct {
		name  string
		index int
	}{
		{
			name:  "negative index",
			index: -1,
		},
		{
			name:  "index equal to length",
			index: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			path := filepath.Join(t.TempDir(), "config.json")
			writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`)
			store := config.NewStore(path)

			// Act
			err := store.Update(context.Background(), tc.index, port.ConfigEntry{
				Name:   "primary",
				DBPath: "/tmp/primary.sqlite",
			})

			// Assert
			if !errors.Is(err, config.ErrDatabaseIndexOutOfRange) {
				t.Fatalf("expected error %v, got %v", config.ErrDatabaseIndexOutOfRange, err)
			}
			assertStoredEntries(t, path, []port.ConfigEntry{
				{Name: "local", DBPath: "/tmp/local.sqlite"},
			})
		})
	}
}

func TestStore_UpdateReturnsErrorWhenSerializedConfigExceedsSizeLimit(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`)
	store := config.NewStore(path)
	baseConfig := config.Config{
		Databases: []config.DatabaseConfig{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	entry := replacementConfigEntryForSerializedSize(t, baseConfig, 0, 1<<20, false)

	// Act
	err := store.Update(context.Background(), 0, entry)

	// Assert
	if !errors.Is(err, config.ErrConfigTooLarge) {
		t.Fatalf("expected error %v, got %v", config.ErrConfigTooLarge, err)
	}
	assertStoredEntries(t, path, []port.ConfigEntry{
		{Name: "local", DBPath: "/tmp/local.sqlite"},
	})
}

func TestStore_DeleteRemovesEntryAtIndex(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"},{"name":"analytics","db_path":"/tmp/analytics.sqlite"}]}`)
	store := config.NewStore(path)

	// Act
	err := store.Delete(context.Background(), 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertStoredEntries(t, path, []port.ConfigEntry{
		{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
	})
}

func TestStore_DeleteReturnsErrorForIndexOutOfRange(t *testing.T) {
	testCases := []struct {
		name  string
		index int
	}{
		{
			name:  "negative index",
			index: -1,
		},
		{
			name:  "index equal to length",
			index: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			path := filepath.Join(t.TempDir(), "config.json")
			writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`)
			store := config.NewStore(path)

			// Act
			err := store.Delete(context.Background(), tc.index)

			// Assert
			if !errors.Is(err, config.ErrDatabaseIndexOutOfRange) {
				t.Fatalf("expected error %v, got %v", config.ErrDatabaseIndexOutOfRange, err)
			}
			assertStoredEntries(t, path, []port.ConfigEntry{
				{Name: "local", DBPath: "/tmp/local.sqlite"},
			})
		})
	}
}

func TestStore_DeleteAllowsRemovingLastEntry(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`)
	store := config.NewStore(path)

	// Act
	err := store.Delete(context.Background(), 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertStoredEntries(t, path, nil)
}

func TestStore_ActivePathReturnsConfiguredPath(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(path)

	// Act
	got, err := store.ActivePath(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != path {
		t.Fatalf("expected %q, got %q", path, got)
	}
}

func configEntryForSerializedSize(t *testing.T, base config.Config, targetSize int, exact bool) port.ConfigEntry {
	t.Helper()

	const dbPath = "/tmp/local.sqlite"

	entry := port.ConfigEntry{
		Name:   "a",
		DBPath: dbPath,
	}

	currentSize := serializedConfigSize(t, appendConfigEntry(base, entry))
	if currentSize > targetSize {
		t.Fatalf("base config already exceeds target size: %d > %d", currentSize, targetSize)
	}

	padding := targetSize - currentSize
	if !exact {
		padding++
	}
	entry.Name += strings.Repeat("a", padding)

	finalSize := serializedConfigSize(t, appendConfigEntry(base, entry))
	if exact && finalSize != targetSize {
		t.Fatalf("expected serialized size %d, got %d", targetSize, finalSize)
	}
	if !exact && finalSize <= targetSize {
		t.Fatalf("expected serialized size above %d, got %d", targetSize, finalSize)
	}

	return entry
}

func appendConfigEntry(base config.Config, entry port.ConfigEntry) config.Config {
	cfg := config.Config{
		Databases: append([]config.DatabaseConfig(nil), base.Databases...),
	}
	cfg.Databases = append(cfg.Databases, config.DatabaseConfig{
		Name: entry.Name,
		Path: entry.DBPath,
	})
	return cfg
}

func replacementConfigEntryForSerializedSize(t *testing.T, base config.Config, index int, targetSize int, exact bool) port.ConfigEntry {
	t.Helper()

	const dbPath = "/tmp/local.sqlite"

	entry := port.ConfigEntry{
		Name:   "a",
		DBPath: dbPath,
	}

	currentSize := serializedConfigSize(t, replaceConfigEntry(base, index, entry))
	if currentSize > targetSize {
		t.Fatalf("base config already exceeds target size: %d > %d", currentSize, targetSize)
	}

	padding := targetSize - currentSize
	if !exact {
		padding++
	}
	entry.Name += strings.Repeat("a", padding)

	finalSize := serializedConfigSize(t, replaceConfigEntry(base, index, entry))
	if exact && finalSize != targetSize {
		t.Fatalf("expected serialized size %d, got %d", targetSize, finalSize)
	}
	if !exact && finalSize <= targetSize {
		t.Fatalf("expected serialized size above %d, got %d", targetSize, finalSize)
	}

	return entry
}

func replaceConfigEntry(base config.Config, index int, entry port.ConfigEntry) config.Config {
	cfg := config.Config{
		Databases: append([]config.DatabaseConfig(nil), base.Databases...),
	}
	cfg.Databases[index] = config.DatabaseConfig{
		Name: entry.Name,
		Path: entry.DBPath,
	}
	return cfg
}

func serializedConfigSize(t *testing.T, cfg config.Config) int {
	t.Helper()

	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to serialize config: %v", err)
	}
	return len(content) + len("\n")
}

func assertStoredEntries(t *testing.T, path string, want []port.ConfigEntry) {
	t.Helper()

	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("expected stored config to be readable, got %v", err)
	}

	got := make([]port.ConfigEntry, len(cfg.Databases))
	for index, database := range cfg.Databases {
		got[index] = port.ConfigEntry{
			Name:   database.Name,
			DBPath: database.Path,
		}
	}

	assertConfigEntries(t, got, want)
}

func assertConfigEntries(t *testing.T, got []port.ConfigEntry, want []port.ConfigEntry) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(got))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("expected entry at index %d to be %#v, got %#v", index, want[index], got[index])
		}
	}
}
