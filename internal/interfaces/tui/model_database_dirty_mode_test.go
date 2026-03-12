package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mgierok/dbc/internal/application/dto"
)

type spySaveDatabaseChangesUseCase struct {
	lastChanges []dto.NamedTableChanges
	err         error
}

func (s *spySaveDatabaseChangesUseCase) ExecuteDTO(ctx context.Context, changes []dto.NamedTableChanges) error {
	s.lastChanges = append([]dto.NamedTableChanges(nil), changes...)
	return s.err
}

func TestSetTableSelection_WithDirtyTablesSwitchesImmediatelyAndPreservesPerTableStaging(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
		staging: testDatabaseStaging(
			map[string]stagingState{
				"users": {
					pendingInserts: []pendingInsertRow{{}},
				},
				"orders": {
					pendingInserts: []pendingInsertRow{{}},
				},
			},
		),
	}

	model.setTableSelection(1)

	if model.overlay.confirmPopup.active {
		t.Fatal("expected table switch without dirty popup")
	}
	if model.read.selectedTable != 1 {
		t.Fatalf("expected selected table 1, got %d", model.read.selectedTable)
	}
	if model.totalRecordRows() != 1 {
		t.Fatalf("expected orders staging to stay available after switch, got %d rows", model.totalRecordRows())
	}

	model.setTableSelection(0)

	if model.read.selectedTable != 0 {
		t.Fatalf("expected selected table 0 after switching back, got %d", model.read.selectedTable)
	}
	if model.totalRecordRows() != 1 {
		t.Fatalf("expected users staging to be restored after switching back, got %d rows", model.totalRecordRows())
	}
}

func TestHandleKey_DirtyQuitCommandOpensDecisionPrompt(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			viewMode:      ViewRecords,
			tables:        []dto.Table{{Name: "users"}},
			selectedTable: 0,
		},
		staging: testDatabaseStaging(
			map[string]stagingState{
				"users": {pendingInserts: []pendingInsertRow{{}}},
			},
		),
	}

	model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	for _, r := range "quit" {
		model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("expected dirty :quit to wait for explicit decision")
		}
	}
	if !model.overlay.confirmPopup.active {
		t.Fatal("expected dirty :quit decision popup to open")
	}
	if model.overlay.confirmPopup.title != "Quit" {
		t.Fatalf("expected Quit popup title, got %q", model.overlay.confirmPopup.title)
	}
}

func TestConfirmSaveChanges_SubmitsAllDirtyTableChanges(t *testing.T) {
	saveChanges := &spySaveDatabaseChangesUseCase{}
	model := &Model{
		ctx:         context.Background(),
		saveChanges: saveChanges,
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 1,
		},
		staging: testDatabaseStaging(
			map[string]stagingState{
				"users": {
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT", Nullable: false},
						},
					},
					pendingInserts: []pendingInsertRow{
						{
							values: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "1", Raw: int64(1)}},
								1: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
							},
							explicitAuto: map[int]bool{},
						},
					},
				},
				"orders": {
					schema: dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "status", Type: "TEXT", Nullable: false},
						},
					},
					pendingUpdates: map[string]recordEdits{
						"id=9": {
							identity: dto.RecordIdentity{
								Keys: []dto.RecordIdentityKey{{Column: "id", Value: dto.StagedValue{Text: "9", Raw: int64(9)}}},
							},
							changes: map[int]stagedEdit{
								1: {Value: dto.StagedValue{Text: "paid", Raw: "paid"}},
							},
						},
					},
				},
			},
		),
	}

	_, cmd := model.confirmSaveChanges()
	if cmd == nil {
		t.Fatal("expected save command")
	}
	msg := cmd()

	result, ok := msg.(saveChangesMsg)
	if !ok {
		t.Fatalf("expected saveChangesMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("expected no error, got %v", result.err)
	}
	if len(saveChanges.lastChanges) != 2 {
		t.Fatalf("expected two dirty table batches, got %d", len(saveChanges.lastChanges))
	}
	if saveChanges.lastChanges[0].TableName != "orders" && saveChanges.lastChanges[1].TableName != "orders" {
		t.Fatalf("expected orders changes to be included, got %+v", saveChanges.lastChanges)
	}
	if saveChanges.lastChanges[0].TableName != "users" && saveChanges.lastChanges[1].TableName != "users" {
		t.Fatalf("expected users changes to be included, got %+v", saveChanges.lastChanges)
	}
}

func TestRenderStatus_ShowsGlobalDirtyCountAcrossTables(t *testing.T) {
	model := &Model{
		staging: testDatabaseStaging(
			map[string]stagingState{
				"users": {
					pendingUpdates: map[string]recordEdits{
						"id=1": {
							changes: map[int]stagedEdit{
								0: {Value: dto.StagedValue{Text: "alice", Raw: "alice"}},
								1: {Value: dto.StagedValue{Text: "admin", Raw: "admin"}},
							},
						},
					},
				},
				"orders": {
					pendingInserts: []pendingInsertRow{{}},
					pendingDeletes: map[string]recordDelete{"id=2": {}},
				},
			},
		),
	}

	status := stripANSI(model.renderStatus(120))

	if !strings.Contains(status, "WRITE (dirty: 3)") {
		t.Fatalf("expected global dirty count in status, got %q", status)
	}
}

func TestRenderTables_ShowsDirtyMarkerForEachDirtyTable(t *testing.T) {
	model := &Model{
		read: runtimeReadState{
			focus:         FocusTables,
			selectedTable: 0,
			tables: []dto.Table{
				{Name: "users"},
				{Name: "orders"},
				{Name: "audit_log"},
			},
		},
		staging: testDatabaseStaging(
			map[string]stagingState{
				"users":  {pendingInserts: []pendingInsertRow{{}}},
				"orders": {pendingDeletes: map[string]recordDelete{"id=2": {}}},
			},
		),
	}

	lines := strings.Join(model.renderTables(24, 4), "\n")

	if !strings.Contains(lines, "users "+dirtyTableMarker) {
		t.Fatalf("expected dirty marker for users, got %q", lines)
	}
	if !strings.Contains(lines, "orders "+dirtyTableMarker) {
		t.Fatalf("expected dirty marker for orders, got %q", lines)
	}
	if strings.Contains(lines, "audit_log "+dirtyTableMarker) {
		t.Fatalf("expected clean table without dirty marker, got %q", lines)
	}
}
