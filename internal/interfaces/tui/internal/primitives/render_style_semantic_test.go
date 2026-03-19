package primitives

import "testing"

func TestRenderStyles_Render_MapsEverySemanticRoleInNormalAndBackdropVariants(t *testing.T) {
	// Arrange
	normal := NewRenderStyles(true)
	backdrop := normal.Backdrop()
	cases := []struct {
		role             SemanticRole
		expectedNormal   string
		expectedBackdrop string
	}{
		{role: SemanticRoleBody, expectedNormal: "text", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleMuted, expectedNormal: "\x1b[2mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleTitle, expectedNormal: "\x1b[1mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleHeader, expectedNormal: "\x1b[1mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleSummary, expectedNormal: "\x1b[1mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleLabel, expectedNormal: "\x1b[1mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleDirty, expectedNormal: "\x1b[1mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleError, expectedNormal: "\x1b[1;4mtext\x1b[0m", expectedBackdrop: "\x1b[2;4mtext\x1b[0m"},
		{role: SemanticRoleSelected, expectedNormal: "\x1b[7mtext\x1b[0m", expectedBackdrop: "\x1b[2mtext\x1b[0m"},
		{role: SemanticRoleDeleted, expectedNormal: "\x1b[9mtext\x1b[0m", expectedBackdrop: "\x1b[2;9mtext\x1b[0m"},
		{role: SemanticRoleSelectedDeleted, expectedNormal: "\x1b[7;9mtext\x1b[0m", expectedBackdrop: "\x1b[2;9mtext\x1b[0m"},
	}

	for _, tc := range cases {
		// Act
		gotNormal := normal.Render(tc.role, "text")
		gotBackdrop := backdrop.Render(tc.role, "text")

		// Assert
		if gotNormal != tc.expectedNormal {
			t.Fatalf("expected normal role %v to render %q, got %q", tc.role, tc.expectedNormal, gotNormal)
		}
		if gotBackdrop != tc.expectedBackdrop {
			t.Fatalf("expected backdrop role %v to render %q, got %q", tc.role, tc.expectedBackdrop, gotBackdrop)
		}
	}
}

func TestRenderStyles_RenderLine_ComposesSemanticSpansWithoutStyleLeakage(t *testing.T) {
	// Arrange
	line := SemanticLine{
		Span(SemanticRoleHeader, "Table"),
		Span(SemanticRoleBody, ": users"),
		Span(SemanticRoleError, " failed"),
	}

	// Act
	rendered := NewRenderStyles(true).RenderLine(line)

	// Assert
	expected := "\x1b[1mTable\x1b[0m: users\x1b[1;4m failed\x1b[0m"
	if rendered != expected {
		t.Fatalf("expected semantic line %q, got %q", expected, rendered)
	}
	if TextWidth(rendered) != TextWidth("Table: users failed") {
		t.Fatalf("expected semantic line width to match plain text, got %d", TextWidth(rendered))
	}
}
