package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
)

type InputKind int

const (
	InputText InputKind = iota
	InputSelect
)

type InputSpec struct {
	Kind    InputKind
	Options []string
}

var (
	ErrNullNotAllowed = errors.New("null value not allowed")
	ErrInvalidValue   = errors.New("invalid value")
)

func InputSpecForType(columnType string) InputSpec {
	if isBooleanType(columnType) {
		return InputSpec{Kind: InputSelect, Options: []string{"true", "false"}}
	}
	if options := parseEnumOptions(columnType); len(options) > 0 {
		return InputSpec{Kind: InputSelect, Options: options}
	}
	return InputSpec{Kind: InputText}
}

func ParseValue(columnType, input string, isNull, nullable bool) (model.Value, error) {
	if isNull {
		if !nullable {
			return model.Value{}, ErrNullNotAllowed
		}
		return model.Value{IsNull: true, Text: "NULL"}, nil
	}

	trimmed := strings.TrimSpace(input)
	switch affinityForType(columnType) {
	case affinityBoolean:
		normalized, ok := parseBoolean(trimmed)
		if !ok {
			return model.Value{}, fmt.Errorf("invalid boolean value: %w", ErrInvalidValue)
		}
		return model.Value{Text: strings.ToLower(trimmed), Raw: normalized}, nil
	case affinityInteger:
		if trimmed == "" {
			return model.Value{}, fmt.Errorf("invalid integer value: %w", ErrInvalidValue)
		}
		typed, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return model.Value{}, fmt.Errorf("invalid integer value: %w", ErrInvalidValue)
		}
		return model.Value{Text: trimmed, Raw: typed}, nil
	case affinityReal:
		if trimmed == "" {
			return model.Value{}, fmt.Errorf("invalid real value: %w", ErrInvalidValue)
		}
		typed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return model.Value{}, fmt.Errorf("invalid real value: %w", ErrInvalidValue)
		}
		return model.Value{Text: trimmed, Raw: typed}, nil
	case affinityNumeric:
		if trimmed == "" {
			return model.Value{}, fmt.Errorf("invalid numeric value: %w", ErrInvalidValue)
		}
		if strings.ContainsAny(trimmed, ".eE") {
			typed, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return model.Value{}, fmt.Errorf("invalid numeric value: %w", ErrInvalidValue)
			}
			return model.Value{Text: trimmed, Raw: typed}, nil
		}
		typed, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return model.Value{}, fmt.Errorf("invalid numeric value: %w", ErrInvalidValue)
		}
		return model.Value{Text: trimmed, Raw: typed}, nil
	case affinityBlob:
		if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
			decoded, err := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(trimmed, "0x"), "0X"))
			if err != nil {
				return model.Value{}, fmt.Errorf("invalid blob value: %w", ErrInvalidValue)
			}
			return model.Value{Text: input, Raw: decoded}, nil
		}
		return model.Value{Text: input, Raw: []byte(input)}, nil
	default:
		return model.Value{Text: input, Raw: input}, nil
	}
}

type valueAffinity int

const (
	affinityText valueAffinity = iota
	affinityInteger
	affinityReal
	affinityNumeric
	affinityBlob
	affinityBoolean
)

func affinityForType(columnType string) valueAffinity {
	normalized := strings.ToUpper(strings.TrimSpace(columnType))
	if normalized == "" {
		return affinityBlob
	}
	if isBooleanType(normalized) {
		return affinityBoolean
	}
	if strings.Contains(normalized, "INT") {
		return affinityInteger
	}
	if strings.Contains(normalized, "CHAR") || strings.Contains(normalized, "CLOB") || strings.Contains(normalized, "TEXT") ||
		strings.Contains(normalized, "DATE") || strings.Contains(normalized, "TIME") ||
		strings.Contains(normalized, "JSON") || strings.Contains(normalized, "UUID") || strings.Contains(normalized, "GUID") {
		return affinityText
	}
	if strings.Contains(normalized, "BLOB") {
		return affinityBlob
	}
	if strings.Contains(normalized, "REAL") || strings.Contains(normalized, "FLOA") || strings.Contains(normalized, "DOUB") {
		return affinityReal
	}
	return affinityNumeric
}

func isBooleanType(columnType string) bool {
	return strings.Contains(strings.ToUpper(columnType), "BOOL")
}

func parseBoolean(input string) (int64, bool) {
	switch strings.ToLower(input) {
	case "true", "1":
		return 1, true
	case "false", "0":
		return 0, true
	default:
		return 0, false
	}
}

func parseEnumOptions(columnType string) []string {
	upper := strings.ToUpper(columnType)
	enumIndex := strings.Index(upper, "ENUM")
	if enumIndex == -1 {
		return nil
	}
	start := strings.Index(columnType[enumIndex:], "(")
	if start == -1 {
		return nil
	}
	start = enumIndex + start + 1
	end := strings.LastIndex(columnType, ")")
	if end == -1 || end <= start {
		return nil
	}
	content := columnType[start:end]
	rawOptions := strings.Split(content, ",")
	options := make([]string, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option := strings.TrimSpace(raw)
		option = strings.Trim(option, `"'`)
		if option == "" {
			continue
		}
		options = append(options, option)
	}
	return options
}
