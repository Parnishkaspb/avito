package database

import "context"

type IDatabase interface {
	RunDatabase(ctx context.Context) error
}
