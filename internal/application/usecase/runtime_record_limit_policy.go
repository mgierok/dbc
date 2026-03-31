package usecase

import "fmt"

const (
	defaultRuntimeRecordLimit = 20
	maxRuntimeRecordLimit     = 1000
)

type RuntimeRecordLimitPolicy struct{}

func NewRuntimeRecordLimitPolicy() *RuntimeRecordLimitPolicy {
	return &RuntimeRecordLimitPolicy{}
}

func (p *RuntimeRecordLimitPolicy) Default() int {
	return defaultRuntimeRecordLimit
}

func (p *RuntimeRecordLimitPolicy) Validate(limit int) error {
	if limit >= 1 && limit <= maxRuntimeRecordLimit {
		return nil
	}
	return fmt.Errorf("%s", p.InvalidSetLimitHint())
}

func (p *RuntimeRecordLimitPolicy) Effective(limit int) int {
	if limit <= 0 {
		return p.Default()
	}
	if limit > maxRuntimeRecordLimit {
		return maxRuntimeRecordLimit
	}
	return limit
}

func (p *RuntimeRecordLimitPolicy) InvalidSetLimitHint() string {
	return fmt.Sprintf("expected :set limit=<1-%d>", maxRuntimeRecordLimit)
}
