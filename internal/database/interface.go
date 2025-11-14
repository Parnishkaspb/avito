package database

import (
	"context"
	"github.com/Parnishkaspb/avito/internal/models"
)

type DB interface {
	RunDatabase(ctx context.Context) error
	сheckTeam(ctx context.Context, teamName string) (bool, error)
	CreateTeam(ctx context.Context, teamAdd models.RequestTeamAdd) (bool, error)
	GetTeam(ctx context.Context, teamName string) (bool, error)
}
