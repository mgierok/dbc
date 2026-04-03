package tui

import "github.com/mgierok/dbc/internal/application/dto"

func newDirtyTableSwitchModel() *Model {
	return withTestStaging(&Model{
		read: runtimeReadState{
			tables:        []dto.Table{{Name: "users"}, {Name: "orders"}},
			selectedTable: 0,
		},
	}, stagingState{
		pendingInserts: []pendingInsertRow{{}},
		pendingUpdates: map[string]recordEdits{
			"id=1": {changes: map[int]stagedEdit{0: {Value: dto.StagedValue{Text: "x", Raw: "x"}}}},
		},
		pendingDeletes: map[string]recordDelete{"id=2": {}},
	})
}
