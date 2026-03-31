package selector

import (
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
