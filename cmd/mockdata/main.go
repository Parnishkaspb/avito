package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set")
	}

	db, err := sql.Open("pgx", dsn)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	users, err := createUsers(ctx, db, 198)
	if err != nil {
		log.Fatal(err)
	}

	teams, err := createTeams(ctx, db, 20)
	if err != nil {
		log.Fatal(err)
	}

	err = assignUsersToTeams(ctx, db, teams, users)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done.")
}

func createUsers(ctx context.Context, db *sql.DB, n int) ([]string, error) {
	ids := make([]string, 0, n)

	for i := 1; i <= n; i++ {
		name := fmt.Sprintf("User_%03d", i)

		var id string
		err := db.QueryRowContext(ctx,
			`INSERT INTO users (name) VALUES ($1) RETURNING id`,
			name,
		).Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func createTeams(ctx context.Context, db *sql.DB, n int) ([]string, error) {
	ids := make([]string, 0, n)

	for i := 1; i <= n; i++ {
		name := fmt.Sprintf("Team_%02d", i)

		var id string
		err := db.QueryRowContext(ctx,
			`INSERT INTO teams (name) VALUES ($1) RETURNING id`,
			name,
		).Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func assignUsersToTeams(ctx context.Context, db *sql.DB, teams, users []string) error {
	rand.Seed(time.Now().UnixNano())

	for _, userID := range users {
		teamID := teams[rand.Intn(len(teams))]
		isAdmin := rand.Intn(10) == 0

		_, err := db.ExecContext(ctx,
			`INSERT INTO team_members (team_id, user_id, is_admin)
             VALUES ($1, $2, $3)`,
			teamID, userID, isAdmin,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
