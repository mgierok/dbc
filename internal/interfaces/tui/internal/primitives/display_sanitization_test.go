package primitives

import "testing"

func TestSanitizeDisplayText_SingleLineRemovesEscapeSequencesAndControls(t *testing.T) {
	// Arrange
	input := "pre\x1b[31mred\x1b[0m\tmid\r\nline\x1b]2;ignored\a\x00\x7fpost"

	// Act
	got := SanitizeDisplayText(input, DisplaySanitizeSingleLine)

	// Assert
	want := "prered mid linepost"
	if got != want {
		t.Fatalf("expected sanitized single-line text %q, got %q", want, got)
	}
}

func TestSanitizeDisplayText_MultilinePreservesNewlinesAndRemovesEscapeSequences(t *testing.T) {
	// Arrange
	input := "alpha\r\nbeta\t\x1b[32mok\x1b[0m\n\x1b]2;title\aomega\x00"

	// Act
	got := SanitizeDisplayText(input, DisplaySanitizeMultiline)

	// Assert
	want := "alpha\nbeta ok\nomega"
	if got != want {
		t.Fatalf("expected sanitized multiline text %q, got %q", want, got)
	}
}

func TestSpanAndSemanticText_SanitizeDisplayTextByDefault(t *testing.T) {
	// Arrange
	text := "users\x1b[31m\r\n"

	// Act
	span := Span(SemanticRoleHeader, text)
	line := SemanticText(SemanticRoleBody, text)

	// Assert
	if span.Text != "users " {
		t.Fatalf("expected sanitized span text, got %q", span.Text)
	}
	if line.PlainText() != "users " {
		t.Fatalf("expected sanitized semantic text, got %q", line.PlainText())
	}
}

func TestSanitizeDisplayText_PreservesUTF8IconsAndFrameGlyphs(t *testing.T) {
	// Arrange
	input := "⚙ local │ path → /tmp/db.sqlite"

	// Act
	got := SanitizeDisplayText(input, DisplaySanitizeSingleLine)

	// Assert
	if got != input {
		t.Fatalf("expected UTF-8 glyphs to remain unchanged, got %q", got)
	}
	if TextWidth(got) != TextWidth(input) {
		t.Fatalf("expected text width to stay unchanged, got %d want %d", TextWidth(got), TextWidth(input))
	}
}
