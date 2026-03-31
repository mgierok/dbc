package usecase

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/sqliteidentity"
)

type RuntimeDatabaseOptionSource string

const (
	RuntimeDatabaseOptionSourceUnknown RuntimeDatabaseOptionSource = ""
	RuntimeDatabaseOptionSourceConfig  RuntimeDatabaseOptionSource = "config"
	RuntimeDatabaseOptionSourceCLI     RuntimeDatabaseOptionSource = "cli"
)

type RuntimeDatabaseOption struct {
	Name       string
	ConnString string
	Source     RuntimeDatabaseOptionSource
}

type RuntimeDatabaseTransitionKind int

const (
	RuntimeDatabaseTransitionReloadCurrent RuntimeDatabaseTransitionKind = iota + 1
	RuntimeDatabaseTransitionOpenDifferent
)

type RuntimeDatabaseTarget struct {
	Option         RuntimeDatabaseOption
	TransitionKind RuntimeDatabaseTransitionKind
}

type RuntimeDatabaseTargetResolver struct{}

func NewRuntimeDatabaseTargetResolver() *RuntimeDatabaseTargetResolver {
	return &RuntimeDatabaseTargetResolver{}
}

func (r *RuntimeDatabaseTargetResolver) Resolve(current RuntimeDatabaseOption, configuredOptions []RuntimeDatabaseOption, requested RuntimeDatabaseOption) (RuntimeDatabaseTarget, error) {
	trimmedConnString := strings.TrimSpace(requested.ConnString)
	if trimmedConnString == "" {
		if strings.TrimSpace(current.ConnString) == "" {
			return RuntimeDatabaseTarget{}, fmt.Errorf("current database unavailable")
		}
		return RuntimeDatabaseTarget{
			Option:         current,
			TransitionKind: RuntimeDatabaseTransitionReloadCurrent,
		}, nil
	}

	resolvedOption := requested
	resolvedOption.ConnString = trimmedConnString
	if strings.TrimSpace(resolvedOption.Name) == "" {
		resolvedOption.Name = trimmedConnString
	}
	if resolvedOption.Source == RuntimeDatabaseOptionSourceUnknown {
		resolvedOption.Source = RuntimeDatabaseOptionSourceCLI
	}

	matchedConfiguredOption := false
	requestedConfiguredOption, requestedConfigured := resolveExactConfiguredRuntimeDatabaseOption(requested, configuredOptions)
	switch {
	case requestedConfigured:
		resolvedOption = requestedConfiguredOption
		matchedConfiguredOption = true
	case true:
		if matched, ok := resolveConfiguredRuntimeDatabaseOption(trimmedConnString, configuredOptions); ok {
			resolvedOption = matched
			matchedConfiguredOption = true
		}
	}
	currentConfiguredOption, currentStillConfigured := resolveExactConfiguredRuntimeDatabaseOption(current, configuredOptions)

	target := RuntimeDatabaseTarget{
		Option:         resolvedOption,
		TransitionKind: RuntimeDatabaseTransitionOpenDifferent,
	}
	if sqliteidentity.Equivalent(resolvedOption.ConnString, current.ConnString) {
		if strings.TrimSpace(current.ConnString) != "" {
			switch {
			case requestedConfigured:
				target.Option = requestedConfiguredOption
			case currentStillConfigured:
				target.Option = currentConfiguredOption
			case !matchedConfiguredOption:
				target.Option = current
			}
		}
		target.TransitionKind = RuntimeDatabaseTransitionReloadCurrent
	}
	return target, nil
}

func resolveConfiguredRuntimeDatabaseOption(connString string, configuredOptions []RuntimeDatabaseOption) (RuntimeDatabaseOption, bool) {
	normalizedConnString := sqliteidentity.Normalize(connString)
	if normalizedConnString == "" {
		return RuntimeDatabaseOption{}, false
	}

	for _, option := range configuredOptions {
		if sqliteidentity.Equivalent(normalizedConnString, option.ConnString) {
			return option, true
		}
	}

	return RuntimeDatabaseOption{}, false
}

func resolveExactConfiguredRuntimeDatabaseOption(requested RuntimeDatabaseOption, configuredOptions []RuntimeDatabaseOption) (RuntimeDatabaseOption, bool) {
	if requested.Source != RuntimeDatabaseOptionSourceConfig || strings.TrimSpace(requested.Name) == "" {
		return RuntimeDatabaseOption{}, false
	}

	normalizedConnString := sqliteidentity.Normalize(requested.ConnString)
	if normalizedConnString == "" {
		return RuntimeDatabaseOption{}, false
	}

	for _, option := range configuredOptions {
		if option.Name != requested.Name {
			continue
		}
		if option.Source != RuntimeDatabaseOptionSourceConfig {
			continue
		}
		if sqliteidentity.Equivalent(normalizedConnString, option.ConnString) {
			return option, true
		}
	}

	return RuntimeDatabaseOption{}, false
}
