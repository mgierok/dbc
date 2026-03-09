package tui

import (
	"context"
	"errors"
	"reflect"
	"testing"

	selectorpkg "github.com/mgierok/dbc/internal/interfaces/tui/internal/selector"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestSelectDatabaseWithState_ReturnsErrorWhenConfigUseCasesMissing(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	_, err := SelectDatabaseWithState(ctx, nil, nil, nil, nil, nil, SelectorLaunchState{})

	// Assert
	if err == nil {
		t.Fatal("expected selector config management validation error")
	}
	if err.Error() != "selector config management use cases are required" {
		t.Fatalf("expected selector config management validation error, got %v", err)
	}
}

func TestSelectDatabaseWithState_DelegatesStateAndUsesConfigManagementAdapter(t *testing.T) {
	// Arrange
	expectedState := SelectorLaunchState{
		StatusMessage:    "Connection failed: ping failed",
		PreferConnString: "/tmp/local.sqlite",
		AdditionalOptions: []DatabaseOption{
			{
				Name:       "/tmp/direct.sqlite",
				ConnString: "/tmp/direct.sqlite",
				Source:     DatabaseOptionSourceCLI,
			},
		},
	}
	expectedOption := DatabaseOption{
		Name:       "warehouse",
		ConnString: "/tmp/warehouse.sqlite",
		Source:     DatabaseOptionSourceConfig,
	}
	store := &fakeSelectorConfigStore{
		entries: []port.ConfigEntry{
			{Name: "local", DBPath: "/tmp/local.sqlite"},
		},
		activePath: "/tmp/config.toml",
	}
	checker := &fakeSelectorConnectionChecker{}
	originalSelector := selectDatabaseWithStateFn
	t.Cleanup(func() {
		selectDatabaseWithStateFn = originalSelector
	})
	selectDatabaseWithStateFn = func(ctx context.Context, manager selectorpkg.Manager, state SelectorLaunchState) (DatabaseOption, error) {
		if !reflect.DeepEqual(state, expectedState) {
			t.Fatalf("expected selector launch state %+v, got %+v", expectedState, state)
		}

		listed, err := manager.List(ctx)
		if err != nil {
			t.Fatalf("expected list without error, got %v", err)
		}
		expectedListed := []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		}
		if !reflect.DeepEqual(listed, expectedListed) {
			t.Fatalf("expected listed entries %+v, got %+v", expectedListed, listed)
		}

		if err := manager.Create(ctx, dto.ConfigDatabase{Name: "analytics", Path: "/tmp/analytics.sqlite"}); err != nil {
			t.Fatalf("expected create without error, got %v", err)
		}
		if err := manager.Update(ctx, 0, dto.ConfigDatabase{Name: "warehouse", Path: "/tmp/warehouse.sqlite"}); err != nil {
			t.Fatalf("expected update without error, got %v", err)
		}
		if err := manager.Delete(ctx, 1); err != nil {
			t.Fatalf("expected delete without error, got %v", err)
		}

		activePath, err := manager.ActivePath(ctx)
		if err != nil {
			t.Fatalf("expected active path without error, got %v", err)
		}
		if activePath != "/tmp/config.toml" {
			t.Fatalf("expected active path %q, got %q", "/tmp/config.toml", activePath)
		}

		return expectedOption, nil
	}

	listConfiguredDatabases := usecase.NewListConfiguredDatabases(store)
	createConfiguredDatabase := usecase.NewCreateConfiguredDatabase(store, checker)
	updateConfiguredDatabase := usecase.NewUpdateConfiguredDatabase(store, checker)
	deleteConfiguredDatabase := usecase.NewDeleteConfiguredDatabase(store)
	getActiveConfigPath := usecase.NewGetActiveConfigPath(store)

	// Act
	selected, err := SelectDatabaseWithState(
		context.Background(),
		listConfiguredDatabases,
		createConfiguredDatabase,
		updateConfiguredDatabase,
		deleteConfiguredDatabase,
		getActiveConfigPath,
		expectedState,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selected != expectedOption {
		t.Fatalf("expected selected option %+v, got %+v", expectedOption, selected)
	}
	if checker.callCount != 2 {
		t.Fatalf("expected connection checker to run twice, got %d", checker.callCount)
	}
	expectedPaths := []string{"/tmp/analytics.sqlite", "/tmp/warehouse.sqlite"}
	if !reflect.DeepEqual(checker.paths, expectedPaths) {
		t.Fatalf("expected connection checker paths %+v, got %+v", expectedPaths, checker.paths)
	}
	if store.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", store.createCalls)
	}
	if store.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", store.updateCalls)
	}
	if store.deleteCalls != 1 {
		t.Fatalf("expected one delete call, got %d", store.deleteCalls)
	}
}

func TestSelectDatabaseWithState_PropagatesDelegatedError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("selector failed")
	store := &fakeSelectorConfigStore{}
	checker := &fakeSelectorConnectionChecker{}
	originalSelector := selectDatabaseWithStateFn
	t.Cleanup(func() {
		selectDatabaseWithStateFn = originalSelector
	})
	selectDatabaseWithStateFn = func(ctx context.Context, manager selectorpkg.Manager, state SelectorLaunchState) (DatabaseOption, error) {
		return DatabaseOption{}, expectedErr
	}

	listConfiguredDatabases := usecase.NewListConfiguredDatabases(store)
	createConfiguredDatabase := usecase.NewCreateConfiguredDatabase(store, checker)
	updateConfiguredDatabase := usecase.NewUpdateConfiguredDatabase(store, checker)
	deleteConfiguredDatabase := usecase.NewDeleteConfiguredDatabase(store)
	getActiveConfigPath := usecase.NewGetActiveConfigPath(store)

	// Act
	_, err := SelectDatabaseWithState(
		context.Background(),
		listConfiguredDatabases,
		createConfiguredDatabase,
		updateConfiguredDatabase,
		deleteConfiguredDatabase,
		getActiveConfigPath,
		SelectorLaunchState{},
	)

	// Assert
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSelectDatabase_UsesEmptyLaunchState(t *testing.T) {
	// Arrange
	expected := DatabaseOption{Name: "local", ConnString: "/tmp/local.sqlite"}
	store := &fakeSelectorConfigStore{}
	checker := &fakeSelectorConnectionChecker{}
	originalSelector := selectDatabaseWithStateFn
	t.Cleanup(func() {
		selectDatabaseWithStateFn = originalSelector
	})
	selectDatabaseWithStateFn = func(ctx context.Context, manager selectorpkg.Manager, state SelectorLaunchState) (DatabaseOption, error) {
		if !reflect.DeepEqual(state, SelectorLaunchState{}) {
			t.Fatalf("expected empty selector launch state, got %+v", state)
		}
		return expected, nil
	}

	listConfiguredDatabases := usecase.NewListConfiguredDatabases(store)
	createConfiguredDatabase := usecase.NewCreateConfiguredDatabase(store, checker)
	updateConfiguredDatabase := usecase.NewUpdateConfiguredDatabase(store, checker)
	deleteConfiguredDatabase := usecase.NewDeleteConfiguredDatabase(store)
	getActiveConfigPath := usecase.NewGetActiveConfigPath(store)

	// Act
	selected, err := SelectDatabase(
		context.Background(),
		listConfiguredDatabases,
		createConfiguredDatabase,
		updateConfiguredDatabase,
		deleteConfiguredDatabase,
		getActiveConfigPath,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if selected != expected {
		t.Fatalf("expected selected option %+v, got %+v", expected, selected)
	}
}

type fakeSelectorConfigStore struct {
	entries    []port.ConfigEntry
	activePath string

	listErr       error
	createErr     error
	updateErr     error
	deleteErr     error
	activePathErr error

	createCalls int
	updateCalls int
	deleteCalls int
}

func (f *fakeSelectorConfigStore) List(_ context.Context) ([]port.ConfigEntry, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	result := make([]port.ConfigEntry, len(f.entries))
	copy(result, f.entries)
	return result, nil
}

func (f *fakeSelectorConfigStore) Create(_ context.Context, entry port.ConfigEntry) error {
	if f.createErr != nil {
		return f.createErr
	}

	f.createCalls++
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeSelectorConfigStore) Update(_ context.Context, index int, entry port.ConfigEntry) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}

	f.updateCalls++
	f.entries[index] = entry
	return nil
}

func (f *fakeSelectorConfigStore) Delete(_ context.Context, index int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}

	f.deleteCalls++
	f.entries = append(f.entries[:index], f.entries[index+1:]...)
	return nil
}

func (f *fakeSelectorConfigStore) ActivePath(_ context.Context) (string, error) {
	if f.activePathErr != nil {
		return "", f.activePathErr
	}

	return f.activePath, nil
}

type fakeSelectorConnectionChecker struct {
	callCount int
	paths     []string
	err       error
}

func (f *fakeSelectorConnectionChecker) CanConnect(_ context.Context, path string) error {
	f.callCount++
	f.paths = append(f.paths, path)
	if f.err != nil {
		return f.err
	}
	return nil
}
