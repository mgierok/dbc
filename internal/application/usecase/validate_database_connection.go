package usecase

import (
	"context"
	"strings"

	"github.com/mgierok/dbc/internal/application/port"
)

type ValidateDatabaseConnection struct {
	connectionChecker port.DatabaseConnectionChecker
}

func NewValidateDatabaseConnection(connectionChecker port.DatabaseConnectionChecker) *ValidateDatabaseConnection {
	return &ValidateDatabaseConnection{connectionChecker: connectionChecker}
}

func (uc *ValidateDatabaseConnection) Execute(ctx context.Context, connString string) error {
	if uc == nil || uc.connectionChecker == nil {
		return nil
	}
	return uc.connectionChecker.CanConnect(ctx, strings.TrimSpace(connString))
}
