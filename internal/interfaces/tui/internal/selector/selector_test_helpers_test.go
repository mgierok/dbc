package selector

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/application/usecase"
)

func newTestSelectorModel(t *testing.T, manager *fakeSelectorManager, state ...SelectorLaunchState) *databaseSelectorModel {
	t.Helper()

	var launchState SelectorLaunchState
	if len(state) > 0 {
		launchState = state[0]
	}

	model, err := newDatabaseSelectorModel(context.Background(), manager, launchState)
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}

	return model
}

func typeText(model *databaseSelectorModel, text string) *databaseSelectorModel {
	current := model
	for _, r := range text {
		current = sendKey(current, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return current
}

func sendKey(model *databaseSelectorModel, key tea.KeyMsg) *databaseSelectorModel {
	updated, _ := model.Update(key)
	return updated.(*databaseSelectorModel)
}

type fakeSelectorManager struct {
	entries    []dto.ConfigDatabase
	activePath string

	loadStateErr  error
	activePathErr error
	createErr     error
	updateErr     error
	deleteErr     error

	loadState     *dto.DatabaseSelectorState
	lastLoadInput dto.DatabaseSelectorLoadInput
	created       []dto.ConfigDatabase
	updated       []updatedEntry
	deleted       []int
}

type updatedEntry struct {
	index int
	entry dto.ConfigDatabase
}

func (f *fakeSelectorManager) LoadState(ctx context.Context, input dto.DatabaseSelectorLoadInput) (dto.DatabaseSelectorState, error) {
	f.lastLoadInput = input
	if f.loadStateErr != nil {
		return dto.DatabaseSelectorState{}, f.loadStateErr
	}
	if f.loadState != nil {
		return *f.loadState, nil
	}

	entries := make([]port.ConfigEntry, len(f.entries))
	for i, entry := range f.entries {
		entries[i] = port.ConfigEntry{Name: entry.Name, DBPath: entry.Path}
	}

	store := fakeSelectorManagerConfigStore{
		entries:       entries,
		activePath:    f.activePath,
		activePathErr: f.activePathErr,
	}
	return usecase.NewLoadDatabaseSelectorState(&store).Execute(ctx, input)
}

func (f *fakeSelectorManager) Create(_ context.Context, entry dto.ConfigDatabase) error {
	if f.createErr != nil {
		return f.createErr
	}
	if strings.TrimSpace(entry.Name) == "" || strings.TrimSpace(entry.Path) == "" {
		return errors.New("invalid entry")
	}
	f.created = append(f.created, entry)
	f.entries = append(f.entries, entry)
	return nil
}

func (f *fakeSelectorManager) Update(_ context.Context, index int, entry dto.ConfigDatabase) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.updated = append(f.updated, updatedEntry{index: index, entry: entry})
	f.entries[index] = entry
	return nil
}

func (f *fakeSelectorManager) Delete(_ context.Context, index int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if index < 0 || index >= len(f.entries) {
		return errors.New("index out of range")
	}
	f.deleted = append(f.deleted, index)
	f.entries = append(f.entries[:index], f.entries[index+1:]...)
	return nil
}

type fakeSelectorManagerConfigStore struct {
	entries       []port.ConfigEntry
	activePath    string
	activePathErr error
}

func (f *fakeSelectorManagerConfigStore) List(_ context.Context) ([]port.ConfigEntry, error) {
	result := make([]port.ConfigEntry, len(f.entries))
	copy(result, f.entries)
	return result, nil
}

func (f *fakeSelectorManagerConfigStore) Create(_ context.Context, _ port.ConfigEntry) error {
	return errors.New("unexpected create call")
}

func (f *fakeSelectorManagerConfigStore) Update(_ context.Context, _ int, _ port.ConfigEntry) error {
	return errors.New("unexpected update call")
}

func (f *fakeSelectorManagerConfigStore) Delete(_ context.Context, _ int) error {
	return errors.New("unexpected delete call")
}

func (f *fakeSelectorManagerConfigStore) ActivePath(_ context.Context) (string, error) {
	if f.activePathErr != nil {
		return "", f.activePathErr
	}
	if f.activePath == "" {
		return "/tmp/config.json", nil
	}
	return f.activePath, nil
}
