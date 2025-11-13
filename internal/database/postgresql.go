package database

import (
	"context"
	"fmt"
	"github.com/Parnishkaspb/avito/internal/config"
	"github.com/jackc/pgx/v5"
	"log"
)

type Database struct {
	User     string
	Password string
	Host     string
	Port     int
	DB       string
	SSLMode  string
}

func New(databaseConfig config.PostreSQLConfig) IDatabase {
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

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("проблемы с подключением к БД: %w", err)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("ошибка! БД не пигнуется: %w", err)
	}

	log.Println("Соединение с БД успешно выполненно")

	<-ctx.Done()
	log.Println("Успешное выключение БД")

	return nil
}
