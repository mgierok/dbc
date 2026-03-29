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

func (r *RuntimeDatabaseTargetResolver) Resolve(current RuntimeDatabaseOption, configuredOptions []RuntimeDatabaseOption, connString string) (RuntimeDatabaseTarget, error) {
	trimmedConnString := strings.TrimSpace(connString)
	if trimmedConnString == "" {
		if strings.TrimSpace(current.ConnString) == "" {
			return RuntimeDatabaseTarget{}, fmt.Errorf("current database unavailable")
		}
		return RuntimeDatabaseTarget{
			Option:         current,
			TransitionKind: RuntimeDatabaseTransitionReloadCurrent,
		}, nil
	}

	resolvedOption := RuntimeDatabaseOption{
		Name:       trimmedConnString,
		ConnString: trimmedConnString,
		Source:     RuntimeDatabaseOptionSourceCLI,
	}
	if matched, ok := resolveConfiguredRuntimeDatabaseOption(trimmedConnString, configuredOptions); ok {
		resolvedOption = matched
	}

	target := RuntimeDatabaseTarget{
		Option:         resolvedOption,
		TransitionKind: RuntimeDatabaseTransitionOpenDifferent,
	}
	if sqliteidentity.Equivalent(resolvedOption.ConnString, current.ConnString) {
		if strings.TrimSpace(current.ConnString) != "" {
			target.Option = current
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
