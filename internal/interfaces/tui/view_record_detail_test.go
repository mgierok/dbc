package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func TestRenderRecordDetail_UsesVerticalLayoutWithoutTruncation(t *testing.T) {
	// Arrange
	longValue := "abcdefghijklmnopqrstuvwxyz0123456789"
	model := newRecordDetailModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER"},
				{Name: "payload", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"1", longValue}}},
	)

	// Act
	lines := model.renderContent(20, 8)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, primitives.IconInfo+" Persisted record") {
		t.Fatalf("expected information marker in detail layout, got %q", content)
	}
	if strings.Contains(content, "[ROW]") {
		t.Fatalf("expected detail layout to render user-facing labels only, got %q", content)
	}
	if !strings.Contains(content, "\x1b[1mid\x1b[0m (INTEGER)") {
		t.Fatalf("expected id header in detail layout, got %q", content)
	}
	if !strings.Contains(content, "\x1b[1mpayload\x1b[0m (TEXT)") {
		t.Fatalf("expected payload header in detail layout, got %q", content)
	}
	if strings.Contains(content, "[COL]") {
		t.Fatalf("expected detail layout without literal column marker tokens, got %q", content)
	}
	if strings.Contains(content, "...") {
		t.Fatalf("expected no truncation marker in detail layout, got %q", content)
	}
}

func TestRecordDetailContentLines_UsesInformationMarkerForRowStates(t *testing.T) {
	for _, tc := range []struct {
		name           string
		model          *Model
		expectedMarker string
		expectedBody   string
	}{
		{
			name: "persisted row",
			model: newStyledRecordsViewModel(
				dto.Schema{Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER", PrimaryKey: true}}},
				[]dto.RecordRow{{Values: []string{"1"}}},
			),
			expectedMarker: primitives.IconInfo + " Persisted record",
		},
		{
			name: "pending insert row",
			model: withTestStaging(newRecordsViewModel(
				dto.Schema{Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER"}}},
				nil,
			), stagingState{
				pendingInserts: []pendingInsertRow{{}},
			}),
			expectedMarker: primitives.IconInfo + " Pending insert",
		},
		{
			name: "delete-marked persisted row",
			model: func() *Model {
				model := newRecordsViewModel(
					dto.Schema{Columns: []dto.SchemaColumn{{Name: "id", Type: "INTEGER", PrimaryKey: true}}},
					[]dto.RecordRow{{Values: []string{"1"}}},
				)
				setTestPendingDeletes(model, map[string]recordDelete{
					mustPersistedRecordKey(t, model, 0): {},
				})
				return model
			}(),
			expectedMarker: primitives.IconInfo + " Marked for delete",
		},
		{
			name: "edited persisted row",
			model: func() *Model {
				model := newStyledRecordsViewModel(
					dto.Schema{
						Columns: []dto.SchemaColumn{
							{Name: "id", Type: "INTEGER", PrimaryKey: true},
							{Name: "name", Type: "TEXT"},
						},
					},
					[]dto.RecordRow{{Values: []string{"1", "alice"}}},
				)
				setTestPendingUpdates(model, map[string]recordEdits{
					mustPersistedRecordKey(t, model, 0): {
						changes: map[int]stagedEdit{
							1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
						},
					},
				})
				return model
			}(),
			expectedMarker: primitives.IconInfo + " Edited record",
			expectedBody:   "\x1b[1mname\x1b[0m (TEXT) " + primitives.IconEdit,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			lines := tc.model.recordDetailContentLines(40)
			content := strings.Join(lines, "\n")

			// Assert
			if !strings.Contains(lines[0], tc.expectedMarker) {
				t.Fatalf("expected row marker %q, got %q", tc.expectedMarker, lines[0])
			}
			if tc.expectedBody != "" && !strings.Contains(content, tc.expectedBody) {
				t.Fatalf("expected detail body fragment %q, got %q", tc.expectedBody, content)
			}
			if tc.name == "edited persisted row" && strings.Contains(content, "\x1b[1mid\x1b[0m (INTEGER) "+primitives.IconEdit) {
				t.Fatalf("expected no edit icon on unmodified field header in detail content, got %q", content)
			}
		})
	}
}

func TestRecordDetailContentLines_ShowsProjectedMetadataBadgesBelowFieldHeader(t *testing.T) {
	// Arrange
	model := newRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{
					Name:           "id",
					Type:           "INTEGER",
					MetadataBadges: []string{"PK", "NOT NULL", "AUTOINCREMENT", "FK->accounts.owner_id"},
				},
				{
					Name:           "nickname",
					Type:           "TEXT",
					MetadataBadges: []string{"NULL", "DEFAULT 'guest'"},
				},
			},
		},
		[]dto.RecordRow{{Values: []string{"1", "alice"}}},
	)

	// Act
	content := stripANSI(strings.Join(model.recordDetailContentLines(80), "\n"))

	// Assert
	if !strings.Contains(content, "id (INTEGER)\n[PK] [NOT NULL] [AUTOINCREMENT] [FK->accounts.owner_id]\n  1") {
		t.Fatalf("expected projected metadata badges under id header, got %q", content)
	}
	if !strings.Contains(content, "nickname (TEXT)\n[NULL] [DEFAULT 'guest']\n  alice") {
		t.Fatalf("expected projected metadata badges under nickname header, got %q", content)
	}
}

func TestRecordDetailContentLines_StrikesDeleteMarkedFieldValuesOnly(t *testing.T) {
	// Arrange
	model := newStyledRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "name", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"1", "alice"}}},
	)
	key := mustPersistedRecordKey(t, model, 0)
	setTestPendingUpdates(model, map[string]recordEdits{
		key: {
			changes: map[int]stagedEdit{
				1: {Value: dto.StagedValue{Text: "alice2", Raw: "alice2"}},
			},
		},
	})
	setTestPendingDeletes(model, map[string]recordDelete{key: {}})

	// Act
	lines := model.recordDetailContentLines(40)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[9malice2\x1b[0m") {
		t.Fatalf("expected delete-marked detail value lines to use strikethrough, got %q", content)
	}
	if strings.Contains(lines[0], "\x1b[9m") {
		t.Fatalf("expected delete summary line to remain readable without strikethrough, got %q", lines[0])
	}
	if strings.Contains(content, "\x1b[9m\x1b[1mname\x1b[0m (TEXT)") || strings.Contains(content, "\x1b[1mname\x1b[0m (TEXT)\x1b[9m") {
		t.Fatalf("expected detail field header to remain unstruck, got %q", content)
	}
}

func TestRecordDetailContentLines_StrikesEveryWrappedDeleteMarkedValueLine(t *testing.T) {
	// Arrange
	model := newStyledRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "notes", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"1", "abcdefghijklmnopqrstuvwx"}}},
	)
	setTestPendingDeletes(model, map[string]recordDelete{
		mustPersistedRecordKey(t, model, 0): {},
	})

	// Act
	lines := model.recordDetailContentLines(14)
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[9mabcdefghijkl\x1b[0m") {
		t.Fatalf("expected first wrapped delete-marked value line to use strikethrough, got %q", content)
	}
	if !strings.Contains(content, "  \x1b[9mmnopqrstuvwx\x1b[0m") {
		t.Fatalf("expected continuation delete-marked value line to keep strikethrough, got %q", content)
	}
	if !strings.Contains(stripANSI(content), "  abcdefghijkl\n  mnopqrstuvwx") {
		t.Fatalf("expected wrapped delete-marked value lines in plain text, got %q", stripANSI(content))
	}
}

func TestRecordDetailContentLinesWithStyles_BackdropKeepsDeletedStyleAcrossWrappedLines(t *testing.T) {
	// Arrange
	model := newRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "notes", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"1", "abcdefghijklmnopqrstuvwx"}}},
	)
	setTestPendingDeletes(model, map[string]recordDelete{
		mustPersistedRecordKey(t, model, 0): {},
	})

	// Act
	lines := model.recordDetailContentLinesWithStyles(14, primitives.NewRenderStyles(true).Backdrop())
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[2;9mabcdefghijkl\x1b[0m") {
		t.Fatalf("expected first backdrop delete-marked value line to use faint strikethrough, got %q", content)
	}
	if !strings.Contains(content, "  \x1b[2;9mmnopqrstuvwx\x1b[0m") {
		t.Fatalf("expected backdrop continuation line to keep faint strikethrough, got %q", content)
	}
	if !strings.Contains(stripANSI(content), "  abcdefghijkl\n  mnopqrstuvwx") {
		t.Fatalf("expected wrapped backdrop delete-marked value lines in plain text, got %q", stripANSI(content))
	}
}

func TestRecordDetailContentLinesWithStyles_BackdropKeepsDeletedStyleAcrossExplicitLineBreaks(t *testing.T) {
	// Arrange
	model := newRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "id", Type: "INTEGER", PrimaryKey: true},
				{Name: "notes", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"1", "alpha\nbeta"}}},
	)
	setTestPendingDeletes(model, map[string]recordDelete{
		mustPersistedRecordKey(t, model, 0): {},
	})

	// Act
	lines := model.recordDetailContentLinesWithStyles(20, primitives.NewRenderStyles(true).Backdrop())
	content := strings.Join(lines, "\n")

	// Assert
	if !strings.Contains(content, "  \x1b[2;9malpha\x1b[0m") {
		t.Fatalf("expected first delete-marked line to use faint strikethrough, got %q", content)
	}
	if !strings.Contains(content, "  \x1b[2;9mbeta\x1b[0m") {
		t.Fatalf("expected line after explicit break to keep faint strikethrough, got %q", content)
	}
}

func TestDeleteMarkedViews_FallBackToPlainTextWhenStylesAreDisabled(t *testing.T) {
	// Arrange
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")

	model := &Model{
		styles: primitives.ResolveRenderStylesFromEnv(),
		read: runtimeReadState{
			viewMode: ViewRecords,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{Name: "id", Type: "INTEGER", PrimaryKey: true},
					{Name: "name", Type: "TEXT"},
				},
			},
			records: []dto.RecordRow{{Values: []string{"1", "alice"}}},
		},
	}
	setTestPendingDeletes(model, map[string]recordDelete{
		mustPersistedRecordKey(t, model, 0): {},
	})

	// Act
	recordContent := strings.Join(model.renderRecords(80, 6), "\n")
	detailLines := model.recordDetailContentLines(40)
	detailContent := strings.Join(detailLines, "\n")

	// Assert
	if strings.Contains(recordContent, "\x1b[") || strings.Contains(detailContent, "\x1b[") {
		t.Fatalf("expected plain-text fallback without ANSI styling, got records=%q detail=%q", recordContent, detailContent)
	}
	if !strings.Contains(recordContent, primitives.IconDelete+" ") {
		t.Fatalf("expected delete marker to remain visible without ANSI styling, got %q", recordContent)
	}
	if !strings.Contains(detailLines[0], primitives.IconInfo+" Marked for delete") {
		t.Fatalf("expected detail delete summary to remain visible without ANSI styling, got %q", detailLines[0])
	}
}
