package tui

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
)

type stagingUIState struct {
	showAuto map[dto.InsertDraftID]bool
}

func displayValue(value dto.StagedValue) string {
	if value.IsNull {
		return "NULL"
	}
	if strings.TrimSpace(value.Text) != "" {
		return value.Text
	}
	if value.Raw != nil {
		return fmt.Sprint(value.Raw)
	}
	return ""
}
