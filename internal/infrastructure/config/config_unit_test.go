package config_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestDecode_ReturnsConfigForValidDocuments(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []config.DatabaseConfig
	}{
		{
			name:  "trimmed empty document",
			input: " \n\t ",
			want:  nil,
		},
		{
			name:  "empty databases list",
			input: `{"databases":[]}`,
			want:  []config.DatabaseConfig{},
		},
		{
			name:  "single database",
			input: `{"databases":[{"name":"local","db_path":"/tmp/example.sqlite"}]}`,
			want: []config.DatabaseConfig{
				{Name: "local", Path: "/tmp/example.sqlite"},
			},
		},
		{
			name:  "multiple databases",
			input: `{"databases":[{"name":"local","db_path":"/tmp/example.sqlite"},{"name":"analytics","db_path":"/tmp/analytics.sqlite"}]}`,
			want: []config.DatabaseConfig{
				{Name: "local", Path: "/tmp/example.sqlite"},
				{Name: "analytics", Path: "/tmp/analytics.sqlite"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := strings.NewReader(tc.input)

			// Act
			got, err := config.Decode(input)

			// Assert
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			assertDatabaseConfigs(t, got.Databases, tc.want)
		})
	}
}

func TestDecode_ReturnsErrorForInvalidDocuments(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		wantErr         error
		wantErrContains string
	}{
		{
			name:            "unknown top level field",
			input:           `{"title":"dbc"}`,
			wantErrContains: `unknown field "title"`,
		},
		{
			name:            "legacy database section",
			input:           `{"database":{"name":"legacy","db_path":"/tmp/example.sqlite"}}`,
			wantErrContains: `unknown field "database"`,
		},
		{
			name:            "multiple json documents",
			input:           `{} {}`,
			wantErrContains: "single JSON object",
		},
		{
			name:    "missing database name",
			input:   `{"databases":[{"db_path":"/tmp/example.sqlite"}]}`,
			wantErr: config.ErrMissingDatabaseName,
		},
		{
			name:    "missing database path",
			input:   `{"databases":[{"name":"local"}]}`,
			wantErr: config.ErrMissingDatabasePath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := strings.NewReader(tc.input)

			// Act
			_, err := config.Decode(input)

			// Assert
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErrContains, err.Error())
			}
		})
	}
}

func TestDecode_EnforcesConfigSizeLimit(t *testing.T) {
	testCases := []struct {
		name    string
		size    int
		wantErr error
	}{
		{
			name: "config at size limit",
			size: 1 << 20,
		},
		{
			name:    "config above size limit",
			size:    (1 << 20) + 1,
			wantErr: config.ErrConfigTooLarge,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			input := configDocumentWithTotalSize(t, tc.size)

			// Act
			cfg, err := config.Decode(strings.NewReader(input))

			// Assert
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(cfg.Databases) != 1 {
				t.Fatalf("expected 1 database, got %d", len(cfg.Databases))
			}
			if cfg.Databases[0].Path != "/tmp/db.sqlite" {
				t.Fatalf("expected path %q, got %q", "/tmp/db.sqlite", cfg.Databases[0].Path)
			}
			if cfg.Databases[0].Name == "" {
				t.Fatal("expected non-empty database name")
			}
		})
	}
}

func TestResolvePathForOS_ReturnsHomeConfigPath(t *testing.T) {
	testCases := []struct {
		name string
		goos string
		home string
	}{
		{
			name: "linux",
			goos: "linux",
			home: "/home/tester",
		},
		{
			name: "macos",
			goos: "darwin",
			home: "/Users/tester",
		},
		{
			name: "unknown os",
			goos: "plan9",
			home: "/home/tester",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			expected := filepath.Join(tc.home, ".config", "dbc", "config.json")

			// Act
			got := config.ResolvePathForOS(tc.goos, tc.home, "")

			// Assert
			if got != expected {
				t.Fatalf("expected %q, got %q", expected, got)
			}
		})
	}
}

func configDocumentWithTotalSize(t *testing.T, totalSize int) string {
	t.Helper()

	const (
		prefix = `{"databases":[{"name":"`
		suffix = `","db_path":"/tmp/db.sqlite"}]}`
	)

	nameLength := totalSize - len(prefix) - len(suffix)
	if nameLength <= 0 {
		t.Fatalf("requested config size %d is too small", totalSize)
	}

	return prefix + strings.Repeat("a", nameLength) + suffix
}

func assertDatabaseConfigs(t *testing.T, got []config.DatabaseConfig, want []config.DatabaseConfig) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d databases, got %d", len(want), len(got))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("expected database at index %d to be %#v, got %#v", index, want[index], got[index])
		}
	}
}
