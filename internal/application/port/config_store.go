package port

import "context"

type ConfigEntry struct {
	Name   string
	DBPath string
}

type ConfigStore interface {
	List(ctx context.Context) ([]ConfigEntry, error)
	Create(ctx context.Context, entry ConfigEntry) error
	Update(ctx context.Context, index int, entry ConfigEntry) error
	Delete(ctx context.Context, index int) error
	ActivePath(ctx context.Context) (string, error)
}
