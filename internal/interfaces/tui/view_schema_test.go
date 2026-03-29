package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func TestRenderSchema_ShowsConstraintBadgesInStableOrder(t *testing.T) {
	// Arrange
	defaultValue := "'guest'"
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewSchema,
			focus:    FocusContent,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{
						Name:           "id",
						Type:           "INTEGER",
						MetadataBadges: []string{"PK", "NOT NULL", "DEFAULT " + defaultValue, "AUTOINCREMENT", "FK->accounts.owner_id", "FK->profiles.id"},
					},
					{
						Name:           "nickname",
						Type:           "TEXT",
						MetadataBadges: []string{"NULL"},
					},
				},
			},
		},
	}

	// Act
	content := stripANSI(strings.Join(model.renderSchema(120, 4), "\n"))

	// Assert
	if !strings.Contains(content, "id : INTEGER [PK] [NOT NULL] [DEFAULT 'guest'] [AUTOINCREMENT] [FK->accounts.owner_id] [FK->profiles.id]") {
		t.Fatalf("expected schema row with ordered badges, got %q", content)
	}
	if !strings.Contains(content, "nickname : TEXT [NULL]") {
		t.Fatalf("expected explicit NULL badge for nullable column, got %q", content)
	}
	if strings.Contains(content, "[UNIQUE]") {
		t.Fatalf("expected primary key column not to duplicate UNIQUE badge, got %q", content)
	}
}

func TestRenderSchema_KeepsRenderedLinesWithinWidth(t *testing.T) {
	// Arrange
	defaultValue := "long-default-value-that-should-be-truncated"
	model := &Model{
		read: runtimeReadState{
			viewMode: ViewSchema,
			schema: dto.Schema{
				Columns: []dto.SchemaColumn{
					{
						Name:         "very_long_column_name",
						Type:         "TEXT",
						Nullable:     false,
						DefaultValue: &defaultValue,
						Unique:       true,
						ForeignKeys: []dto.ForeignKeyRef{
							{Table: "extremely_long_reference_table", Column: "extremely_long_reference_column"},
						},
					},
				},
			},
		},
	}

	// Act
	lines := model.renderSchema(32, 3)

	// Assert
	for _, line := range lines {
		if primitives.TextWidth(stripANSI(line)) > 32 {
			t.Fatalf("expected rendered line width <= 32, got %d for %q", primitives.TextWidth(stripANSI(line)), stripANSI(line))
		}
	}
}
