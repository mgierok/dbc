package port

import "context"

type DatabaseConnectionChecker interface {
	CanConnect(ctx context.Context, dbPath string) error
}
