package tui

import (
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
)

func TestRenderRecords_SanitizesHeaderAndRowValues(t *testing.T) {
	// Arrange
	model := newRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "na\x1b[31mme", Type: "TEXT"},
				{Name: "notes", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"ali\tce\x1b[0m", "line1\r\nline2\x00"}}},
	)

	// Act
	content := stripANSI(strings.Join(model.renderRecords(60, 6), "\n"))

	// Assert
	if strings.Contains(content, "\x1b") || strings.Contains(content, "\r") || strings.Contains(content, "\nline2") {
		t.Fatalf("expected records view to strip escape/control characters from headers and cells, got %q", content)
	}
	if !strings.Contains(content, "name") {
		t.Fatalf("expected sanitized header label, got %q", content)
	}
	if !strings.Contains(content, "ali ce") {
		t.Fatalf("expected sanitized cell value, got %q", content)
	}
	if !strings.Contains(content, "line1 line2") {
		t.Fatalf("expected single-line sanitization for tabular record rows, got %q", content)
	}
}

func TestRecordDetailContentLines_PreservesMultilineWhileSanitizingEscapeSequences(t *testing.T) {
	// Arrange
	model := newRecordsViewModel(
		dto.Schema{
			Columns: []dto.SchemaColumn{
				{Name: "payload\x1b[31m", Type: "TEXT"},
			},
		},
		[]dto.RecordRow{{Values: []string{"alpha\r\nbeta\t\x1b[32mok\x1b[0m\nomega\x00"}}},
	)

	// Act
	content := stripANSI(strings.Join(model.recordDetailContentLines(40), "\n"))

	// Assert
	if strings.Contains(content, "\x1b") || strings.Contains(content, "\r") {
		t.Fatalf("expected record detail to strip escape sequences and carriage returns, got %q", content)
	}
	if !strings.Contains(content, "payload (TEXT)") {
		t.Fatalf("expected sanitized column header, got %q", content)
	}
	if !strings.Contains(content, "  alpha\n  beta ok\n  omega") {
		t.Fatalf("expected record detail to preserve multiline content after sanitization, got %q", content)
	}
}
