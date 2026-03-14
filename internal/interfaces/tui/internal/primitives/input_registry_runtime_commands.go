package primitives

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"
)

type RuntimeCommandAction int

const (
	RuntimeCommandActionNone RuntimeCommandAction = iota
	RuntimeCommandActionOpenHelp
	RuntimeCommandActionQuit
	RuntimeCommandActionOpenConfig
	RuntimeCommandActionSave
	RuntimeCommandActionSetRecordLimit
)

type runtimeCommandMatcher func(input string, spec RuntimeCommandSpec) (RuntimeCommandSpec, bool, error)

type RuntimeCommandSpec struct {
	Aliases     []string
	Usage       string
	Description string
	Action      RuntimeCommandAction
	RecordLimit int
	matcher     runtimeCommandMatcher
}

var (
	ErrUnknownRuntimeCommand = errors.New("unknown runtime command")
	errInvalidRuntimeCommand = errors.New("invalid runtime command")
)

var runtimeCommandSpecs = []RuntimeCommandSpec{
	{
		Aliases:     []string{"config", "c"},
		Description: "Open database selector.",
		Action:      RuntimeCommandActionOpenConfig,
	},
	{
		Aliases:     []string{"help", "h"},
		Description: "Open runtime help popup reference.",
		Action:      RuntimeCommandActionOpenHelp,
	},
	{
		Aliases:     []string{"w", "write"},
		Description: "Save staged changes.",
		Action:      RuntimeCommandActionSave,
	},
	{
		Usage:       ":set limit=<n>",
		Description: "Set records page limit for the current app session.",
		Action:      RuntimeCommandActionSetRecordLimit,
		matcher:     matchSetRecordLimitCommand,
	},
	{
		Aliases:     []string{"quit", "q"},
		Description: "Quit the application.",
		Action:      RuntimeCommandActionQuit,
	},
}

func ParseRuntimeCommand(input string) (RuntimeCommandSpec, error) {
	command := normalizeRuntimeCommand(input)
	if command == "" {
		return RuntimeCommandSpec{}, ErrUnknownRuntimeCommand
	}

	for _, candidate := range runtimeCommandSpecs {
		if candidate.matcher != nil {
			matched, ok, err := candidate.matcher(command, candidate)
			if ok {
				return cloneRuntimeCommandSpec(matched), err
			}
			continue
		}

		for _, alias := range candidate.Aliases {
			if strings.EqualFold(command, alias) {
				return cloneRuntimeCommandSpec(candidate), nil
			}
		}
	}

	return RuntimeCommandSpec{}, ErrUnknownRuntimeCommand
}

func normalizeRuntimeCommand(input string) string {
	command := strings.TrimSpace(input)
	command = strings.TrimPrefix(command, ":")
	return strings.TrimSpace(command)
}

func matchSetRecordLimitCommand(input string, spec RuntimeCommandSpec) (RuntimeCommandSpec, bool, error) {
	setKeyword, remainder, matched := splitRuntimeCommandKeyword(input)
	if !matched || !strings.EqualFold(setKeyword, "set") {
		return RuntimeCommandSpec{}, false, nil
	}

	remainder = strings.TrimSpace(remainder)
	if remainder == "" {
		return RuntimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	lowerRemainder := strings.ToLower(remainder)
	if !strings.HasPrefix(lowerRemainder, "limit") {
		return RuntimeCommandSpec{}, false, nil
	}
	if !strings.HasPrefix(lowerRemainder, "limit=") {
		return RuntimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	recordLimitText := remainder[len("limit="):]
	if recordLimitText == "" || strings.ContainsAny(recordLimitText, " \t") {
		return RuntimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	recordLimit, err := strconv.Atoi(recordLimitText)
	if err != nil || !runtimecontract.IsValidRecordPageLimit(recordLimit) {
		return RuntimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	matchedSpec := spec
	matchedSpec.RecordLimit = recordLimit
	return matchedSpec, true, nil
}

func splitRuntimeCommandKeyword(input string) (string, string, bool) {
	keywordEnd := strings.IndexAny(input, " \t")
	if keywordEnd == -1 {
		if strings.TrimSpace(input) == "" {
			return "", "", false
		}
		return input, "", true
	}

	return input[:keywordEnd], input[keywordEnd:], true
}

func invalidSetRecordLimitCommandError() error {
	return fmt.Errorf("%w: %s", errInvalidRuntimeCommand, runtimecontract.InvalidSetRecordLimitHint())
}

func IsUnknownRuntimeCommand(err error) bool {
	return errors.Is(err, ErrUnknownRuntimeCommand)
}

func runtimeCommandLabel(command RuntimeCommandSpec) string {
	if strings.TrimSpace(command.Usage) != "" {
		return command.Usage
	}
	aliases := make([]string, 0, len(command.Aliases))
	for _, alias := range command.Aliases {
		aliases = append(aliases, ":"+alias)
	}
	return strings.Join(aliases, " / ")
}

func runtimeCommandLabelForAction(action RuntimeCommandAction) string {
	for _, command := range runtimeCommandSpecs {
		if command.Action == action {
			return runtimeCommandLabel(command)
		}
	}
	return ""
}

func cloneRuntimeCommandSpec(spec RuntimeCommandSpec) RuntimeCommandSpec {
	spec.Aliases = append([]string(nil), spec.Aliases...)
	return spec
}
