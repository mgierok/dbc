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

func (f *fakeConfigStore) List(ctx context.Context) ([]port.ConfigEntry, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]port.ConfigEntry(nil), f.entries...), nil
}

func (f *fakeConfigStore) Create(ctx context.Context, entry port.ConfigEntry) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createCalls++
	f.lastCreated = entry
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeConfigStore) Update(ctx context.Context, index int, entry port.ConfigEntry) error {
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

func (f *fakeConfigStore) Delete(ctx context.Context, index int) error {
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

func (f *fakeConfigStore) ActivePath(ctx context.Context) (string, error) {
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

func TestListConfiguredDatabases_MapsEntries(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
			{Name: "analytics", DBPath: "/tmp/analytics.sqlite"},
		},
	}
	uc := usecase.NewListConfiguredDatabases(store)

	// Act
	result, err := uc.Execute(context.Background())

	// Assert
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
}

func TestGetActiveConfigPath_ReturnsPath(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{activePath: "/tmp/config.toml"}
	uc := usecase.NewGetActiveConfigPath(store)

	// Act
	path, err := uc.Execute(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != "/tmp/config.toml" {
		t.Fatalf("expected path %q, got %q", "/tmp/config.toml", path)
	}
}

func TestCreateConfiguredDatabase_CreatesEntry(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{}
	checker := &fakeDatabaseConnectionChecker{}
	uc := usecase.NewCreateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), dto.ConfigDatabase{
		Name: "local",
		Path: "/tmp/local.sqlite",
	})

	// Assert
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
}

func TestCreateConfiguredDatabase_ValidatesName(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{}
	checker := &fakeDatabaseConnectionChecker{}
	uc := usecase.NewCreateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), dto.ConfigDatabase{
		Name: "  ",
		Path: "/tmp/local.sqlite",
	})

	// Assert
	if !errors.Is(err, usecase.ErrConfigDatabaseNameRequired) {
		t.Fatalf("expected error %v, got %v", usecase.ErrConfigDatabaseNameRequired, err)
	}
}

func TestCreateConfiguredDatabase_ValidatesPath(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{}
	checker := &fakeDatabaseConnectionChecker{}
	uc := usecase.NewCreateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), dto.ConfigDatabase{
		Name: "local",
		Path: " ",
	})

	// Assert
	if !errors.Is(err, usecase.ErrConfigDatabasePathRequired) {
		t.Fatalf("expected error %v, got %v", usecase.ErrConfigDatabasePathRequired, err)
	}
}

func TestUpdateConfiguredDatabase_UpdatesEntry(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
		},
	}
	checker := &fakeDatabaseConnectionChecker{}
	uc := usecase.NewUpdateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), 0, dto.ConfigDatabase{
		Name: "prod",
		Path: "/tmp/prod.sqlite",
	})

	// Assert
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
}

func TestUpdateConfiguredDatabase_ValidatesIndex(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{}
	checker := &fakeDatabaseConnectionChecker{}
	uc := usecase.NewUpdateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), -1, dto.ConfigDatabase{
		Name: "prod",
		Path: "/tmp/prod.sqlite",
	})

	// Assert
	if !errors.Is(err, usecase.ErrConfigDatabaseIndexOutOfRange) {
		t.Fatalf("expected error %v, got %v", usecase.ErrConfigDatabaseIndexOutOfRange, err)
	}
}

func TestCreateConfiguredDatabase_BlocksCreateWhenConnectionValidationFails(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{}
	checker := &fakeDatabaseConnectionChecker{err: errors.New("cannot connect")}
	uc := usecase.NewCreateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), dto.ConfigDatabase{
		Name: "local",
		Path: " /tmp/local.sqlite ",
	})

	// Assert
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
}

func TestUpdateConfiguredDatabase_BlocksUpdateWhenConnectionValidationFails(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
		},
	}
	checker := &fakeDatabaseConnectionChecker{err: errors.New("cannot connect")}
	uc := usecase.NewUpdateConfiguredDatabase(store, checker)

	// Act
	err := uc.Execute(context.Background(), 0, dto.ConfigDatabase{
		Name: "prod",
		Path: " /tmp/prod.sqlite ",
	})

	// Assert
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
}

func TestDeleteConfiguredDatabase_DeletesEntry(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
			{Name: "prod", DBPath: "/tmp/prod.sqlite"},
		},
	}
	uc := usecase.NewDeleteConfiguredDatabase(store)

	// Act
	err := uc.Execute(context.Background(), 1)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store.lastDeletedIndex != 1 {
		t.Fatalf("expected deleted index %d, got %d", 1, store.lastDeletedIndex)
	}
}

func TestDeleteConfiguredDatabase_ValidatesIndex(t *testing.T) {
	// Arrange
	store := &fakeConfigStore{}
	uc := usecase.NewDeleteConfiguredDatabase(store)

	// Act
	err := uc.Execute(context.Background(), -1)

	// Assert
	if !errors.Is(err, usecase.ErrConfigDatabaseIndexOutOfRange) {
		t.Fatalf("expected error %v, got %v", usecase.ErrConfigDatabaseIndexOutOfRange, err)
	}
}
