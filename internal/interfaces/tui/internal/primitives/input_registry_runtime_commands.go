package primitives

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	runtimecontract "github.com/mgierok/dbc/internal/interfaces/tui/internal"
)

type runtimeCommandAction int

const (
	runtimeCommandActionNone runtimeCommandAction = iota
	runtimeCommandActionOpenHelp
	runtimeCommandActionQuit
	runtimeCommandActionOpenConfig
	runtimeCommandActionSetRecordLimit
)

type runtimeCommandMatcher func(input string, spec runtimeCommandSpec) (runtimeCommandSpec, bool, error)

type runtimeCommandSpec struct {
	aliases     []string
	usage       string
	description string
	action      runtimeCommandAction
	recordLimit int
	matcher     runtimeCommandMatcher
}

var (
	errUnknownRuntimeCommand = errors.New("unknown runtime command")
	errInvalidRuntimeCommand = errors.New("invalid runtime command")
)

var runtimeCommandSpecs = []runtimeCommandSpec{
	{
		aliases:     []string{"config", "c"},
		description: "Open database selector and config manager.",
		action:      runtimeCommandActionOpenConfig,
	},
	{
		aliases:     []string{"help", "h"},
		description: "Open runtime help popup reference.",
		action:      runtimeCommandActionOpenHelp,
	},
	{
		usage:       ":set limit=<n>",
		description: "Set records page limit for the current app session.",
		action:      runtimeCommandActionSetRecordLimit,
		matcher:     matchSetRecordLimitCommand,
	},
	{
		aliases:     []string{"quit", "q"},
		description: "Quit the application.",
		action:      runtimeCommandActionQuit,
	},
}

func resolveRuntimeCommand(input string) (runtimeCommandSpec, bool) {
	spec, err := parseRuntimeCommand(input)
	if err != nil {
		return runtimeCommandSpec{}, false
	}
	return spec, true
}

func parseRuntimeCommand(input string) (runtimeCommandSpec, error) {
	command := normalizeRuntimeCommand(input)
	if command == "" {
		return runtimeCommandSpec{}, errUnknownRuntimeCommand
	}

	for _, candidate := range runtimeCommandSpecs {
		if candidate.matcher != nil {
			matched, ok, err := candidate.matcher(command, candidate)
			if ok {
				return matched, err
			}
			continue
		}

		for _, alias := range candidate.aliases {
			if strings.EqualFold(command, alias) {
				return candidate, nil
			}
		}
	}

	return runtimeCommandSpec{}, errUnknownRuntimeCommand
}

func normalizeRuntimeCommand(input string) string {
	command := strings.TrimSpace(input)
	command = strings.TrimPrefix(command, ":")
	return strings.TrimSpace(command)
}

func matchSetRecordLimitCommand(input string, spec runtimeCommandSpec) (runtimeCommandSpec, bool, error) {
	setKeyword, remainder, matched := splitRuntimeCommandKeyword(input)
	if !matched || !strings.EqualFold(setKeyword, "set") {
		return runtimeCommandSpec{}, false, nil
	}

	remainder = strings.TrimSpace(remainder)
	if remainder == "" {
		return runtimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	lowerRemainder := strings.ToLower(remainder)
	if !strings.HasPrefix(lowerRemainder, "limit") {
		return runtimeCommandSpec{}, false, nil
	}
	if !strings.HasPrefix(lowerRemainder, "limit=") {
		return runtimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	recordLimitText := remainder[len("limit="):]
	if recordLimitText == "" || strings.ContainsAny(recordLimitText, " \t") {
		return runtimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	recordLimit, err := strconv.Atoi(recordLimitText)
	if err != nil || !runtimecontract.IsValidRecordPageLimit(recordLimit) {
		return runtimeCommandSpec{}, true, invalidSetRecordLimitCommandError()
	}

	matchedSpec := spec
	matchedSpec.recordLimit = recordLimit
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

func runtimeCommandLabel(command runtimeCommandSpec) string {
	if strings.TrimSpace(command.usage) != "" {
		return command.usage
	}
	aliases := make([]string, 0, len(command.aliases))
	for _, alias := range command.aliases {
		aliases = append(aliases, ":"+alias)
	}
	return strings.Join(aliases, " / ")
}
