package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
)

type fakeConfigStore struct {
	entries       []port.ConfigEntry
	activePath    string
	listErr       error
	createErr     error
	updateErr     error
	deleteErr     error
	activePathErr error

	lastCreated      port.ConfigEntry
	lastUpdatedIndex int
	lastUpdatedEntry port.ConfigEntry
	lastDeletedIndex int
	createCalls      int
	updateCalls      int
}

func (f *fakeConfigStore) List(context.Context) ([]port.ConfigEntry, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]port.ConfigEntry(nil), f.entries...), nil
}

func (f *fakeConfigStore) Create(_ context.Context, entry port.ConfigEntry) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createCalls++
	f.lastCreated = entry
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeConfigStore) Update(_ context.Context, index int, entry port.ConfigEntry) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updateCalls++
	f.lastUpdatedIndex = index
	f.lastUpdatedEntry = entry
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.entries[index] = entry
	return nil
}

func (f *fakeConfigStore) Delete(_ context.Context, index int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.lastDeletedIndex = index
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.entries = append(f.entries[:index], f.entries[index+1:]...)
	return nil
}

func (f *fakeConfigStore) ActivePath(context.Context) (string, error) {
	if f.activePathErr != nil {
		return "", f.activePathErr
	}
	return f.activePath, nil
}

type fakeDatabaseConnectionChecker struct {
	err       error
	callCount int
	lastPath  string
}

func (f *fakeDatabaseConnectionChecker) CanConnect(_ context.Context, dbPath string) error {
	f.callCount++
	f.lastPath = dbPath
	if f.err != nil {
		return f.err
	}
	return nil
}

func TestConfigManagement_MappingContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "list configured databases maps entries",
			run: func(t *testing.T) {
				t.Helper()

				store := &fakeConfigStore{
					entries: []port.ConfigEntry{
						{Name: "local", DBPath: "/tmp/local.sqlite"},
						{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
					},
				}
				uc := usecase.NewListConfiguredDatabases(store)

				result, err := uc.Execute(context.Background())

				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				expected := []dto.ConfigDatabase{
					{Name: "local", Path: "/tmp/local.sqlite"},
					{Name: "analytics", Path: "/tmp/analytics.sqlite"},
				}
				if !reflect.DeepEqual(result, expected) {
					t.Fatalf("expected %v, got %v", expected, result)
				}
			},
		},
		{
			name: "get active config path returns active path",
			run: func(t *testing.T) {
				t.Helper()

				uc := usecase.NewGetActiveConfigPath(&fakeConfigStore{activePath: "/tmp/config.json"})

				path, err := uc.Execute(context.Background())

				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if path != "/tmp/config.json" {
					t.Fatalf("expected path %q, got %q", "/tmp/config.json", path)
				}
			},
		},
		{
			name: "create configured database trims and persists entry",
			run: func(t *testing.T) {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewCreateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), dto.ConfigDatabase{
					Name: " local ",
					Path: " /tmp/local.sqlite ",
				})

				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				expected := port.ConfigEntry{Name: "local", DBPath: "/tmp/local.sqlite"}
				if !reflect.DeepEqual(store.lastCreated, expected) {
					t.Fatalf("expected created entry %v, got %v", expected, store.lastCreated)
				}
				if checker.callCount != 1 {
					t.Fatalf("expected connection checker call count %d, got %d", 1, checker.callCount)
				}
				if checker.lastPath != "/tmp/local.sqlite" {
					t.Fatalf("expected checker path %q, got %q", "/tmp/local.sqlite", checker.lastPath)
				}
			},
		},
		{
			name: "update configured database trims and persists entry",
			run: func(t *testing.T) {
				t.Helper()

				store := &fakeConfigStore{
					entries: []port.ConfigEntry{
						{Name: "local", DBPath: "/tmp/local.sqlite"},
					},
				}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewUpdateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), 0, dto.ConfigDatabase{
					Name: " prod ",
					Path: " /tmp/prod.sqlite ",
				})

				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if store.lastUpdatedIndex != 0 {
					t.Fatalf("expected index %d, got %d", 0, store.lastUpdatedIndex)
				}

				expected := port.ConfigEntry{Name: "prod", DBPath: "/tmp/prod.sqlite"}
				if !reflect.DeepEqual(store.lastUpdatedEntry, expected) {
					t.Fatalf("expected updated entry %v, got %v", expected, store.lastUpdatedEntry)
				}
				if checker.callCount != 1 {
					t.Fatalf("expected connection checker call count %d, got %d", 1, checker.callCount)
				}
				if checker.lastPath != "/tmp/prod.sqlite" {
					t.Fatalf("expected checker path %q, got %q", "/tmp/prod.sqlite", checker.lastPath)
				}
			},
		},
		{
			name: "delete configured database forwards index to store",
			run: func(t *testing.T) {
				t.Helper()

				store := &fakeConfigStore{
					entries: []port.ConfigEntry{
						{Name: "local", DBPath: "/tmp/local.sqlite"},
						{Name: "prod", DBPath: "/tmp/prod.sqlite"},
					},
				}
				uc := usecase.NewDeleteConfiguredDatabase(store)

				err := uc.Execute(context.Background(), 1)

				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if store.lastDeletedIndex != 1 {
					t.Fatalf("expected deleted index %d, got %d", 1, store.lastDeletedIndex)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}

func TestConfigManagement_ValidationRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T) error
		want error
	}{
		{
			name: "create requires name",
			run: func(t *testing.T) error {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewCreateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), dto.ConfigDatabase{
					Name: "  ",
					Path: "/tmp/local.sqlite",
				})
				if checker.callCount != 0 {
					t.Fatalf("expected no connection validation, got %d calls", checker.callCount)
				}
				if store.createCalls != 0 {
					t.Fatalf("expected no create call, got %d", store.createCalls)
				}
				return err
			},
			want: usecase.ErrConfigDatabaseNameRequired,
		},
		{
			name: "create requires path",
			run: func(t *testing.T) error {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewCreateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), dto.ConfigDatabase{
					Name: "local",
					Path: " ",
				})
				if checker.callCount != 0 {
					t.Fatalf("expected no connection validation, got %d calls", checker.callCount)
				}
				if store.createCalls != 0 {
					t.Fatalf("expected no create call, got %d", store.createCalls)
				}
				return err
			},
			want: usecase.ErrConfigDatabasePathRequired,
		},
		{
			name: "update requires non-negative index",
			run: func(t *testing.T) error {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewUpdateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), -1, dto.ConfigDatabase{
					Name: "prod",
					Path: "/tmp/prod.sqlite",
				})
				if checker.callCount != 0 {
					t.Fatalf("expected no connection validation, got %d calls", checker.callCount)
				}
				if store.updateCalls != 0 {
					t.Fatalf("expected no update call, got %d", store.updateCalls)
				}
				return err
			},
			want: usecase.ErrConfigDatabaseIndexOutOfRange,
		},
		{
			name: "update requires name",
			run: func(t *testing.T) error {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewUpdateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), 0, dto.ConfigDatabase{
					Name: " ",
					Path: "/tmp/prod.sqlite",
				})
				if checker.callCount != 0 {
					t.Fatalf("expected no connection validation, got %d calls", checker.callCount)
				}
				if store.updateCalls != 0 {
					t.Fatalf("expected no update call, got %d", store.updateCalls)
				}
				return err
			},
			want: usecase.ErrConfigDatabaseNameRequired,
		},
		{
			name: "update requires path",
			run: func(t *testing.T) error {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{}
				uc := usecase.NewUpdateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), 0, dto.ConfigDatabase{
					Name: "prod",
					Path: " ",
				})
				if checker.callCount != 0 {
					t.Fatalf("expected no connection validation, got %d calls", checker.callCount)
				}
				if store.updateCalls != 0 {
					t.Fatalf("expected no update call, got %d", store.updateCalls)
				}
				return err
			},
			want: usecase.ErrConfigDatabasePathRequired,
		},
		{
			name: "delete requires non-negative index",
			run: func(t *testing.T) error {
				t.Helper()
				return usecase.NewDeleteConfiguredDatabase(&fakeConfigStore{}).Execute(context.Background(), -1)
			},
			want: usecase.ErrConfigDatabaseIndexOutOfRange,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.run(t)

			if !errors.Is(err, tc.want) {
				t.Fatalf("expected error %v, got %v", tc.want, err)
			}
		})
	}
}

func TestConfigManagement_BlocksPersistenceWhenConnectionValidationFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "create",
			run: func(t *testing.T) {
				t.Helper()

				store := &fakeConfigStore{}
				checker := &fakeDatabaseConnectionChecker{err: errors.New("cannot connect")}
				uc := usecase.NewCreateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), dto.ConfigDatabase{
					Name: "local",
					Path: " /tmp/local.sqlite ",
				})

				if !errors.Is(err, usecase.ErrConfigDatabaseConnectionFailed) {
					t.Fatalf("expected error %v, got %v", usecase.ErrConfigDatabaseConnectionFailed, err)
				}
				if checker.callCount != 1 {
					t.Fatalf("expected checker call count %d, got %d", 1, checker.callCount)
				}
				if checker.lastPath != "/tmp/local.sqlite" {
					t.Fatalf("expected checker path %q, got %q", "/tmp/local.sqlite", checker.lastPath)
				}
				if store.createCalls != 0 {
					t.Fatalf("expected no create call, got %d", store.createCalls)
				}
			},
		},
		{
			name: "update",
			run: func(t *testing.T) {
				t.Helper()

				store := &fakeConfigStore{
					entries: []port.ConfigEntry{
						{Name: "local", DBPath: "/tmp/local.sqlite"},
					},
				}
				checker := &fakeDatabaseConnectionChecker{err: errors.New("cannot connect")}
				uc := usecase.NewUpdateConfiguredDatabase(store, checker)

				err := uc.Execute(context.Background(), 0, dto.ConfigDatabase{
					Name: "prod",
					Path: " /tmp/prod.sqlite ",
				})

				if !errors.Is(err, usecase.ErrConfigDatabaseConnectionFailed) {
					t.Fatalf("expected error %v, got %v", usecase.ErrConfigDatabaseConnectionFailed, err)
				}
				if checker.callCount != 1 {
					t.Fatalf("expected checker call count %d, got %d", 1, checker.callCount)
				}
				if checker.lastPath != "/tmp/prod.sqlite" {
					t.Fatalf("expected checker path %q, got %q", "/tmp/prod.sqlite", checker.lastPath)
				}
				if store.updateCalls != 0 {
					t.Fatalf("expected no update call, got %d", store.updateCalls)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}
