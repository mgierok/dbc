package selector

import (
	"context"
	"testing"

	"github.com/mgierok/dbc/internal/application/dto"
	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func TestNewDatabaseSelectorModel_UsesDetectedRenderStyles(t *testing.T) {
	// Arrange
	originalDetector := detectRenderStyles
	t.Cleanup(func() {
		detectRenderStyles = originalDetector
	})
	detectRenderStyles = func() primitives.RenderStyles {
		return primitives.NewRenderStyles(true)
	}

	manager := &fakeSelectorManager{
		entries: []dto.ConfigDatabase{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}

	// Act
	model, err := newDatabaseSelectorModel(context.Background(), manager)

	// Assert
	if err != nil {
		t.Fatalf("expected selector model, got error %v", err)
	}
	if !model.styles.Enabled() {
		t.Fatal("expected selector model to keep detected render styles")
	}
}
