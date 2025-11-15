package database

import (
	"context"
	"github.com/Parnishkaspb/avito/internal/models"
)

type DB interface {
	RunDatabase(ctx context.Context) error
	CheckTeam(ctx context.Context, teamName string) (bool, error)
	CreateTeam(ctx context.Context, teamAdd models.RequestTeamAdd) (bool, error)
	GetTeam(ctx context.Context, teamName string) (bool, error)
	CheckRoleUser(ctx context.Context, user_id string) (bool, error)
	ReturnTeamID(ctx context.Context, teamName string) (string, bool, error)
	ReturnTeamMembersByTeamID(ctx context.Context, teamID string) ([]models.RequestMembers, error)
}
