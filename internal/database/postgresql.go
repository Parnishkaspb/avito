package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Parnishkaspb/avito/internal/config"
	"github.com/Parnishkaspb/avito/internal/helper"
	"github.com/Parnishkaspb/avito/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strings"
	"time"
)

type Database struct {
	User     string
	Password string
	Host     string
	Port     int
	DB       string
	SSLMode  string
	Pool     *pgxpool.Pool
}

func New(databaseConfig config.PostreSQLConfig) *Database {
	return &Database{
		User:     databaseConfig.User,
		Password: databaseConfig.Password,
		Host:     databaseConfig.Host,
		Port:     databaseConfig.Port,
		DB:       databaseConfig.DB,
		SSLMode:  databaseConfig.SSLMode,
	}
}

func (db *Database) RunDatabase(ctx context.Context) error {
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.DB, db.SSLMode,
	)

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return fmt.Errorf("проблемы с подключением к БД: %w", err)
	}

	db.Pool = pool

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ошибка! БД не пингуется: %w", err)
	}

	log.Println("Соединение с БД успешно выполнено")

	<-ctx.Done()
	log.Println("Закрываем соединение с БД")
	pool.Close()
	return nil
}

func (db *Database) CheckTeam(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM teams WHERE name=$1)",
		teamName,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	return exists, nil
}

func (db *Database) CreateTeam(ctx context.Context, teamAdd models.RequestTeamAddResponse) (bool, error) {
	exists, err := db.CheckTeam(ctx, teamAdd.TeamName)
	if err != nil {
		log.Printf("проблемы с проверкой команды: %s", err)
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	if exists {
		return exists, nil
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var id string
	err = tx.QueryRow(ctx, "INSERT INTO teams(name) VALUES($1) RETURNING id", teamAdd.TeamName).Scan(&id)
	if err != nil {
		return false, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	rows := helper.ParseMembers(id, teamAdd.Members)

	if rows == nil {
		err = fmt.Errorf("не удалось распарсить участников команды")
		return false, err
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"team_members"},
		[]string{"team_id", "user_id"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return false, fmt.Errorf("ошибка CopyFrom: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	return true, nil
}

func (db *Database) GetTeam(ctx context.Context, teamName string) (bool, error) {
	exists, err := db.CheckTeam(ctx, teamName)
	if err != nil {
		log.Printf("проблемы с проверкой команды: %s", err)
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	return exists, nil
}

func (db *Database) CheckExists(ctx context.Context, table, id string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1)", table)

	err := db.Pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	return exists, nil
}

func (db *Database) CheckUser(ctx context.Context, userID string) (bool, error) {
	return db.CheckExists(ctx, "users", userID)
}

func (db *Database) CheckPR(ctx context.Context, prID string) (bool, error) {
	return db.CheckExists(ctx, "pull_requests", prID)
}

func (db *Database) GetUser(ctx context.Context, userID string) (models.UserActiveResponse, error) {
	var user models.UserActiveResponse
	err := db.Pool.QueryRow(
		ctx,
		`SELECT u.id as user_id, u.name as username, COALESCE(t.name, '') as team_name, u.is_active FROM users u 
    		LEFT JOIN team_members tm ON tm.user_id = u.id 
    		LEFT JOIN teams t ON t.id = tm.team_id 
            WHERE u.id = $1;`,
		userID,
	).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		if err == pgx.ErrNoRows {
			return models.UserActiveResponse{}, nil
		}
		return models.UserActiveResponse{}, fmt.Errorf("ошибка базы данных: %w", err)
	}

	return user, nil
}

func (db *Database) UpdateActive(ctx context.Context, userID string, isActive bool) (bool, error) {
	_, err := db.Pool.Exec(ctx, "UPDATE users SET is_active = $1 WHERE id = $2", isActive, userID)
	if err != nil {
		return false, fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return true, nil
}

func (db *Database) CheckRoleUser(ctx context.Context, userID string) (bool, error) {
	var is_admin bool
	err := db.Pool.QueryRow(
		ctx,
		"SELECT is_admin FROM team_members WHERE user_id = $1",
		userID,
	).Scan(&is_admin)

	if err != nil {
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	return is_admin, nil
}

func (db *Database) ReturnTeamID(ctx context.Context, teamName string) (string, bool, error) {
	var teamID string
	err := db.Pool.QueryRow(
		ctx,
		"SELECT id FROM teams WHERE name = $1",
		teamName,
	).Scan(&teamID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("ошибка базы данных: %w", err)
	}

	return teamID, true, nil
}

func (db *Database) ExecuteQuery(ctx context.Context, query string, args []interface{}, processRow func(rows pgx.Rows) error) error {
	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := processRow(rows); err != nil {
			return fmt.Errorf("ошибка чтения данных: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("ошибка обработки результатов: %w", err)
	}

	return nil
}

func (db *Database) ReturnTeamMembersByTeamID(ctx context.Context, teamID string) ([]models.RequestMembers, error) {
	var members []models.RequestMembers

	err := db.ExecuteQuery(
		ctx,
		`SELECT tm.user_id, u.name as username, u.is_active 
         FROM team_members tm 
         INNER JOIN users u ON u.id = tm.user_id 
         WHERE u.is_active = true AND tm.team_id = $1`,
		[]interface{}{teamID},
		func(rows pgx.Rows) error {
			var member models.RequestMembers
			if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
				return err
			}
			members = append(members, member)
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if members == nil {
		members = []models.RequestMembers{}
	}

	return members, nil
}

func (db *Database) ReturnUserReviewByUserID(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	var pullRequests []models.PullRequestShort

	err := db.ExecuteQuery(
		ctx,
		`SELECT pr.id as pull_request_id, pr.name as pull_request_name, pr.author_id, prs.name as status 
         FROM pull_requests pr 
         INNER JOIN pull_request_assigned_reviewers prar ON prar.pull_request_id = pr.id 
         INNER JOIN pull_request_statuses prs ON prs.id = pr.status 
         WHERE prar.user_id=$1`,
		[]interface{}{userID},
		func(rows pgx.Rows) error {
			var pr models.PullRequestShort
			if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
				return err
			}
			pullRequests = append(pullRequests, pr)
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if pullRequests == nil {
		pullRequests = []models.PullRequestShort{}
	}

	return pullRequests, nil
}

func (db *Database) ReturnTeamMembersByUserID(ctx context.Context, userID string) ([]string, error) {
	rows, err := db.Pool.Query(
		ctx,
		`SELECT DISTINCT u.id
		FROM users u
		INNER JOIN team_members tm ON tm.user_id = u.id
		WHERE tm.team_id IN (
			SELECT tm2.team_id
			FROM team_members tm2
			INNER JOIN users u2 ON u2.id = tm2.user_id
			WHERE u2.id = $1 AND u2.is_active = TRUE
		)
		AND u.is_active = TRUE
		AND u.id <> $1;`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения строк: %w", err)
	}

	return userIDs, nil
}

func (db *Database) GetAvailableTeamMatesForPR(ctx context.Context, prID string) ([]string, error) {
	rows, err := db.Pool.Query(
		ctx,
		`SELECT u.id
		 FROM users u
		 INNER JOIN team_members tm ON tm.user_id = u.id
		 WHERE tm.team_id = (
		     SELECT tm2.team_id
		     FROM team_members tm2
		     INNER JOIN pull_requests pr ON pr.author_id = tm2.user_id
		     WHERE pr.id = $1
		 )
		 AND u.is_active = TRUE
		 AND u.id <> (SELECT author_id FROM pull_requests WHERE id = $1)
		 AND u.id NOT IN (
		     SELECT user_id
		     FROM pull_request_assigned_reviewers
		     WHERE pull_request_id = $1
		 );`,
		prID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения строк: %w", err)
	}

	return userIDs, nil
}

func (db *Database) CreatePullRequest(ctx context.Context, id, name, authorID string) (models.PullRequestShort, error) {
	var pr models.PullRequestShort

	_, err := db.Pool.Exec(
		ctx,
		`INSERT INTO pull_requests (id, name, author_id) VALUES ($1, $2, $3)`,
		id, name, authorID,
	)
	if err != nil {
		return models.PullRequestShort{}, fmt.Errorf("ошибка создания PR: %w", err)
	}

	err = db.Pool.QueryRow(
		ctx,
		`SELECT pr.id, pr.name, pr.author_id, prs.name AS status_name
     FROM pull_requests pr
     INNER JOIN pull_request_statuses prs ON pr.status = prs.id
     WHERE pr.id = $1`,
		id,
	).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status)

	if err != nil {
		return models.PullRequestShort{}, fmt.Errorf("ошибка вывода PR: %w", err)
	}

	return pr, nil
}

func (db *Database) CreatePullRequestAssignedReview(ctx context.Context, prID string, reviewerIDs []string) (bool, error) {
	values := make([]string, 0, len(reviewerIDs))
	args := make([]interface{}, 0, len(reviewerIDs)*2)

	argPos := 1
	for _, reviewerID := range reviewerIDs {
		values = append(values, fmt.Sprintf("($%d, $%d)", argPos, argPos+1))
		args = append(args, prID, reviewerID)
		argPos += 2
	}

	query := fmt.Sprintf(
		"INSERT INTO pull_request_assigned_reviewers (pull_request_id, user_id) VALUES %s",
		strings.Join(values, ","),
	)

	_, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("ошибка создания PR: %w", err)
	}

	return true, nil
}

func (db *Database) MergePullRequest(ctx context.Context, prID string) (models.PullRequest, error) {
	var pr models.PullRequest

	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return pr, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	var dbStatus string
	var mergedAt *time.Time

	err = tx.QueryRow(
		ctx,
		`SELECT id, name, author_id, status, merged_at
		 FROM pull_requests
		 WHERE id = $1
		 FOR UPDATE`,
		prID,
	).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &dbStatus, &mergedAt) // сканируем в string
	if err != nil {
		return pr, fmt.Errorf("ошибка при получении PR: %w", err)
	}

	if dbStatus == "1" {
		_, err = tx.Exec(
			ctx,
			"UPDATE pull_requests SET status = '2', merged_at = NOW() WHERE id = $1",
			prID,
		)
		if err != nil {
			return pr, fmt.Errorf("ошибка при слиянии PR: %w", err)
		}
		dbStatus = "MERGED"

		err = tx.QueryRow(
			ctx,
			"SELECT merged_at FROM pull_requests WHERE id = $1",
			prID,
		).Scan(&mergedAt)
		if err != nil {
			return pr, fmt.Errorf("ошибка при получении времени слияния: %w", err)
		}
	}

	pr.Status = dbStatus

	if mergedAt != nil {
		pr.MergedAt = mergedAt
	} else {
		pr.MergedAt = nil
	}

	var reviewers []string
	err = tx.QueryRow(
		ctx,
		`SELECT COALESCE(array_agg(user_id), '{}') 
		 FROM pull_request_assigned_reviewers 
		 WHERE pull_request_id = $1`,
		prID,
	).Scan(&reviewers)
	if err != nil {
		return pr, fmt.Errorf("ошибка при получении ревьюеров: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (db *Database) ReassignPullRequest(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	var assignedID string
	err := db.Pool.QueryRow(
		ctx,
		"SELECT id FROM pull_request_assigned_reviewers WHERE pull_request_id=$1 AND user_id=$2",
		prID, oldReviewerID,
	).Scan(&assignedID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("ошибка запроса: %w", err)
	}

	_, err = db.Pool.Exec(
		ctx,
		"UPDATE pull_request_assigned_reviewers SET user_id=$1 WHERE id=$2",
		newReviewerID, assignedID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления ревьюера: %w", err)
	}

	return nil
}

func (db *Database) CheckStatusPR(ctx context.Context, prID string) (bool, error) {
	var status string

	err := db.Pool.QueryRow(ctx, "SELECT status FROM pull_requests WHERE id=$1", prID).Scan(&status)
	if err != nil {
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	return status == "1", nil
}

func (db *Database) PullRequestFullInformation(ctx context.Context, prID string) (models.PullRequestResponse, error) {
	var pr models.PullRequestResponse
	pr.PullRequestID = prID

	err := db.Pool.QueryRow(
		ctx,
		`
		SELECT pr.name, pr.author_id, prs.name
		FROM pull_requests pr
		INNER JOIN pull_request_statuses prs ON prs.id = pr.status
		WHERE pr.id = $1
		`,
		prID,
	).Scan(&pr.PullRequestName, &pr.AuthorID, &pr.Status)

	if err != nil {
		if err == pgx.ErrNoRows {
			return models.PullRequestResponse{}, fmt.Errorf("PR %s не найден: %w", prID, err)
		}
		return models.PullRequestResponse{}, fmt.Errorf("ошибка при получении PR: %w", err)
	}

	var reviewers []string
	err = db.Pool.QueryRow(
		ctx,
		`SELECT COALESCE(array_agg(user_id), '{}') 
		 FROM pull_request_assigned_reviewers 
		 WHERE pull_request_id = $1`,
		prID,
	).Scan(&reviewers)
	if err != nil {
		return models.PullRequestResponse{}, fmt.Errorf("ошибка при получении ревьюеров: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (db *Database) GetTeamMetrics(ctx context.Context) ([]models.TeamMetrics, int, int, error) {
	var totalPRs, totalTeams int

	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM pull_requests").Scan(&totalPRs)
	if err != nil {
		return nil, 0, 0, err
	}

	err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM teams").Scan(&totalTeams)
	if err != nil {
		return nil, 0, 0, err
	}

	rows, err := db.Pool.Query(ctx, `
        SELECT 
            t.id,
            t.name,
            COALESCE(admins.admin_count,0),
            COALESCE(participants.pr_participants,0)
        FROM teams t
        LEFT JOIN (
            SELECT team_id, COUNT(*) AS admin_count
            FROM team_members
            WHERE is_admin = TRUE
            GROUP BY team_id
        ) admins ON t.id = admins.team_id
        LEFT JOIN (
            SELECT tm.team_id, COUNT(DISTINCT u.id) AS pr_participants
            FROM team_members tm
            JOIN users u ON u.id = tm.user_id
            LEFT JOIN pull_requests pr ON pr.author_id = u.id
            LEFT JOIN pull_request_assigned_reviewers r ON r.user_id = u.id
            WHERE pr.id IS NOT NULL OR r.id IS NOT NULL
            GROUP BY tm.team_id
        ) participants ON t.id = participants.team_id
    `)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	var metrics []models.TeamMetrics
	for rows.Next() {
		var m models.TeamMetrics
		if err := rows.Scan(&m.TeamID, &m.TeamName, &m.AdminCount, &m.PRParticipants); err != nil {
			return nil, 0, 0, err
		}
		metrics = append(metrics, m)
	}

	return metrics, totalPRs, totalTeams, nil
}
