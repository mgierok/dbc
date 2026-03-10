package selector

import (
	"context"
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

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

	listErr   error
	createErr error
	updateErr error
	deleteErr error

	created []dto.ConfigDatabase
	updated []updatedEntry
	deleted []int
}

type updatedEntry struct {
	index int
	entry dto.ConfigDatabase
}

func (f *fakeSelectorManager) List(_ context.Context) ([]dto.ConfigDatabase, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	result := make([]dto.ConfigDatabase, len(f.entries))
	copy(result, f.entries)
	return result, nil
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

func (f *fakeSelectorManager) ActivePath(_ context.Context) (string, error) {
	if f.activePath == "" {
		return "/tmp/config.toml", nil
	}
	return f.activePath, nil
}
