package runtimecontract

import "fmt"

const MaxRecordPageLimit = 1000

func IsValidRecordPageLimit(limit int) bool {
	return limit >= 1 && limit <= MaxRecordPageLimit
}

func InvalidSetRecordLimitHint() string {
	return fmt.Sprintf("expected :set limit=<1-%d>", MaxRecordPageLimit)
}
