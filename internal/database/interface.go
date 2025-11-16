package database

import (
	"context"
	"github.com/Parnishkaspb/avito/internal/models"
	"github.com/jackc/pgx/v5"
)

type DB interface {
	RunDatabase(ctx context.Context) error
	CheckTeam(ctx context.Context, teamName string) (bool, error)
	CreateTeam(ctx context.Context, teamAdd models.RequestTeamAddResponse) (bool, error)
	GetTeam(ctx context.Context, teamName string) (bool, error)
	CheckRoleUser(ctx context.Context, userID string) (bool, error)
	ReturnTeamID(ctx context.Context, teamName string) (string, bool, error)
	ReturnTeamMembersByTeamID(ctx context.Context, teamID string) ([]models.RequestMembers, error)
	CheckExists(ctx context.Context, table, id string) (bool, error)
	CheckUser(ctx context.Context, userID string) (bool, error)
	CheckPR(ctx context.Context, prID string) (bool, error)
	ReturnTeamMembersByUserID(ctx context.Context, userID string) ([]string, error)
	CreatePullRequest(ctx context.Context, id, name, authorID string) (models.PullRequestShort, error)
	GetUser(ctx context.Context, userID string) (models.UserActiveResponse, error)
	UpdateActive(ctx context.Context, userID string, isActive bool) (bool, error)
	ReturnUserReviewByUserID(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	CreatePullRequestAssignedReview(ctx context.Context, prID string, reviewerIDs []string) (bool, error)
	ExecuteQuery(ctx context.Context, query string, args []interface{}, processRow func(rows pgx.Rows) error) error
	MergePullRequest(ctx context.Context, prID string) (models.PullRequest, error)
	ReassignPullRequest(ctx context.Context, prID, oldReviewerID, newReviewID string) error
	CheckStatusPR(ctx context.Context, prID string) (bool, error)
	GetAvailableTeamMatesForPR(ctx context.Context, prID string) ([]string, error)
	PullRequestFullInformation(ctx context.Context, prID string) (models.PullRequestResponse, error)
	GetTeamMetrics(ctx context.Context) ([]models.TeamMetrics, int, int, error)
}
