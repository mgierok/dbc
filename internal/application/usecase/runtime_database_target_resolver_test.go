package usecase_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mgierok/dbc/internal/application/usecase"
)

func TestRuntimeDatabaseTargetResolver_Resolve_BlankConnStringReloadsCurrentDatabase(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	current := usecase.RuntimeDatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(current, nil, usecase.RuntimeDatabaseOption{ConnString: "   "})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionReloadCurrent {
		t.Fatalf("expected reload-current transition, got %v", target.TransitionKind)
	}
	if target.Option != current {
		t.Fatalf("expected current database target %+v, got %+v", current, target.Option)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_EquivalentConnStringReloadsCurrentDatabase(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	currentPath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := usecase.RuntimeDatabaseOption{
		Name:       "primary",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(current, nil, runtimeDatabaseRequestOption(currentPath+string(os.PathSeparator)+"."))

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionReloadCurrent {
		t.Fatalf("expected reload-current transition, got %v", target.TransitionKind)
	}
	if target.Option.ConnString != current.ConnString {
		t.Fatalf("expected current conn string %q, got %q", current.ConnString, target.Option.ConnString)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_EquivalentConfiguredTargetPreservesConfiguredIdentityOnReload(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	currentPath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := usecase.RuntimeDatabaseOption{
		Name:       currentPath,
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceCLI,
	}
	configured := usecase.RuntimeDatabaseOption{
		Name:       "primary",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(
		current,
		[]usecase.RuntimeDatabaseOption{configured},
		runtimeDatabaseRequestOption(currentPath+string(os.PathSeparator)+"."),
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionReloadCurrent {
		t.Fatalf("expected reload-current transition, got %v", target.TransitionKind)
	}
	if target.Option != configured {
		t.Fatalf("expected configured identity %+v, got %+v", configured, target.Option)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_EquivalentConfiguredReloadPreservesCurrentAliasWhenAnotherConfiguredAliasMatchesFirst(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	currentPath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := usecase.RuntimeDatabaseOption{
		Name:       "primary",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}
	otherAlias := usecase.RuntimeDatabaseOption{
		Name:       "primary-copy",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(
		current,
		[]usecase.RuntimeDatabaseOption{otherAlias, current},
		runtimeDatabaseRequestOption(currentPath+string(os.PathSeparator)+"."),
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionReloadCurrent {
		t.Fatalf("expected reload-current transition, got %v", target.TransitionKind)
	}
	if target.Option != current {
		t.Fatalf("expected current configured alias %+v, got %+v", current, target.Option)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_EquivalentConfiguredReloadUsesLiveConfiguredAliasWhenCurrentIdentityIsStale(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	currentPath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := usecase.RuntimeDatabaseOption{
		Name:       "primary-old",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}
	liveConfigured := usecase.RuntimeDatabaseOption{
		Name:       "primary-renamed",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(
		current,
		[]usecase.RuntimeDatabaseOption{liveConfigured},
		runtimeDatabaseRequestOption(currentPath+string(os.PathSeparator)+"."),
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionReloadCurrent {
		t.Fatalf("expected reload-current transition, got %v", target.TransitionKind)
	}
	if target.Option != liveConfigured {
		t.Fatalf("expected live configured alias %+v, got %+v", liveConfigured, target.Option)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_ExplicitConfiguredAliasSelectionWinsForEquivalentReload(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	currentPath := filepath.Join(t.TempDir(), "primary.sqlite")
	current := usecase.RuntimeDatabaseOption{
		Name:       "primary",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}
	selectedAlias := usecase.RuntimeDatabaseOption{
		Name:       "primary-copy",
		ConnString: currentPath,
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(
		current,
		[]usecase.RuntimeDatabaseOption{current, selectedAlias},
		selectedAlias,
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionReloadCurrent {
		t.Fatalf("expected reload-current transition, got %v", target.TransitionKind)
	}
	if target.Option != selectedAlias {
		t.Fatalf("expected selected configured alias %+v, got %+v", selectedAlias, target.Option)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_ConfiguredMatchWinsOverAnonymousCLIEquivalent(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	configuredPath := filepath.Join(t.TempDir(), "analytics.sqlite")

	// Act
	target, err := resolver.Resolve(
		usecase.RuntimeDatabaseOption{},
		[]usecase.RuntimeDatabaseOption{
			{
				Name:       "analytics",
				ConnString: configuredPath,
				Source:     usecase.RuntimeDatabaseOptionSourceConfig,
			},
		},
		runtimeDatabaseRequestOption(configuredPath+string(os.PathSeparator)+"."),
	)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.Option.Name != "analytics" {
		t.Fatalf("expected configured option name %q, got %q", "analytics", target.Option.Name)
	}
	if target.Option.Source != usecase.RuntimeDatabaseOptionSourceConfig {
		t.Fatalf("expected configured source, got %q", target.Option.Source)
	}
}

func TestRuntimeDatabaseTargetResolver_Resolve_DistinctConnStringOpensDifferentDatabase(t *testing.T) {
	// Arrange
	resolver := usecase.NewRuntimeDatabaseTargetResolver()
	current := usecase.RuntimeDatabaseOption{
		Name:       "primary",
		ConnString: "/tmp/primary.sqlite",
		Source:     usecase.RuntimeDatabaseOptionSourceConfig,
	}

	// Act
	target, err := resolver.Resolve(current, nil, runtimeDatabaseRequestOption("/tmp/analytics.sqlite"))

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.TransitionKind != usecase.RuntimeDatabaseTransitionOpenDifferent {
		t.Fatalf("expected open-different transition, got %v", target.TransitionKind)
	}
	if target.Option.ConnString != "/tmp/analytics.sqlite" {
		t.Fatalf("expected target conn string %q, got %q", "/tmp/analytics.sqlite", target.Option.ConnString)
	}
	if target.Option.Source != usecase.RuntimeDatabaseOptionSourceCLI {
		t.Fatalf("expected CLI source for distinct conn string, got %q", target.Option.Source)
	}
}

func runtimeDatabaseRequestOption(connString string) usecase.RuntimeDatabaseOption {
	return usecase.RuntimeDatabaseOption{
		Name:       connString,
		ConnString: connString,
		Source:     usecase.RuntimeDatabaseOptionSourceCLI,
	}
}
