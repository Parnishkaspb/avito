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

func (db *Database) Run(ctx context.Context) error {
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.DB, db.SSLMode,
	)

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("✅ Database connected successfully")

	<-ctx.Done()
	log.Println("✅ Database shutdown completed")

	return nil
}
