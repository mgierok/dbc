package sqliteidentity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalize_ReturnsAbsoluteCleanPathForRelativeInput(t *testing.T) {
	t.Parallel()

	// Arrange
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected working directory, got %v", err)
	}
	absolutePath := filepath.Join(t.TempDir(), "runtime.sqlite")
	relativePath, err := filepath.Rel(cwd, absolutePath)
	if err != nil {
		t.Fatalf("expected relative path, got %v", err)
	}

	// Act
	normalized := Normalize(relativePath + string(os.PathSeparator) + ".")

	// Assert
	if normalized != absolutePath {
		t.Fatalf("expected normalized path %q, got %q", absolutePath, normalized)
	}
}

func TestEquivalent_ReturnsTrueForEquivalentPaths(t *testing.T) {
	t.Parallel()

	// Arrange
	absolutePath := filepath.Join(t.TempDir(), "runtime.sqlite")
	equivalentPath := absolutePath + string(os.PathSeparator) + "."

	// Act
	equivalent := Equivalent(absolutePath, equivalentPath)

	// Assert
	if !equivalent {
		t.Fatalf("expected paths %q and %q to be equivalent", absolutePath, equivalentPath)
	}
}
