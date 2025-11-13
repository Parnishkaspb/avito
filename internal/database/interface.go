package database

import "context"

type IDatabase interface {
	Run(ctx context.Context) error
}
