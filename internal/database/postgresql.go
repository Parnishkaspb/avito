package database

import (
	"context"
	"fmt"
	"github.com/Parnishkaspb/avito/internal/config"
	"github.com/Parnishkaspb/avito/internal/helper"
	"github.com/Parnishkaspb/avito/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
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
	defer pool.Close()

	db.Pool = pool

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ошибка! БД не пингуется: %w", err)
	}

	log.Println("Соединение с БД успешно выполнено")

	<-ctx.Done()
	log.Println("Успешное выключение БД")
	return nil
}

func (db *Database) сheckTeam(ctx context.Context, teamName string) (bool, error) {
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

func (db *Database) CreateTeam(ctx context.Context, teamAdd models.RequestTeamAdd) (bool, error) {
	exists, err := db.сheckTeam(ctx, teamAdd.TeamName)
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
		[]string{"team_id", "user_id", "is_active"},
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
	exists, err := db.сheckTeam(ctx, teamName)
	if err != nil {
		log.Printf("проблемы с проверкой команды: %s", err)
		return false, fmt.Errorf("ошибка запроса: %w", err)
	}

	return exists, nil
}
