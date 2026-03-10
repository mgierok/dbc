package selector

import (
	"strings"

	"github.com/mgierok/dbc/internal/interfaces/tui/internal/primitives"
)

func (o DatabaseOption) source() DatabaseOptionSource {
	if o.Source == DatabaseOptionSourceCLI {
		return DatabaseOptionSourceCLI
	}
	return DatabaseOptionSourceConfig
}

func (o DatabaseOption) marker() string {
	if o.source() == DatabaseOptionSourceCLI {
		return primitives.IconCLISource
	}
	return primitives.IconConfigSource
}

func (o DatabaseOption) isConfigBacked() bool {
	return o.source() == DatabaseOptionSourceConfig && o.managerIndex >= 0
}

func normalizeAdditionalOptions(options []DatabaseOption) []DatabaseOption {
	if len(options) == 0 {
		return nil
	}

	normalized := make([]DatabaseOption, 0, len(options))
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		connString := strings.TrimSpace(option.ConnString)
		if connString == "" {
			continue
		}
		if _, exists := seen[connString]; exists {
			continue
		}
		seen[connString] = struct{}{}

		name := strings.TrimSpace(option.Name)
		if name == "" {
			name = connString
		}
		normalized = append(normalized, DatabaseOption{
			Name:         name,
			ConnString:   connString,
			Source:       DatabaseOptionSourceCLI,
			managerIndex: -1,
		})
	}
	return normalized
}

func mergeConfigAndAdditionalOptions(configOptions []DatabaseOption, additionalOptions []DatabaseOption) []DatabaseOption {
	if len(additionalOptions) == 0 {
		return configOptions
	}

	merged := make([]DatabaseOption, 0, len(configOptions)+len(additionalOptions))
	merged = append(merged, configOptions...)

	seen := make(map[string]struct{}, len(configOptions)+len(additionalOptions))
	for _, option := range configOptions {
		key := strings.TrimSpace(option.ConnString)
		if key == "" {
			continue
		}
		seen[key] = struct{}{}
	}

	for _, option := range additionalOptions {
		key := strings.TrimSpace(option.ConnString)
		if key != "" {
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
		}
		merged = append(merged, option)
	}

	return merged
}
