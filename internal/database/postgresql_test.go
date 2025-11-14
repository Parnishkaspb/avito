package database

import (
	"github.com/Parnishkaspb/avito/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	cfg := config.PostreSQLConfig{
		User:     "postgres",
		Password: "secret",
		Host:     "localhost",
		Port:     5432,
		DB:       "testdb",
		SSLMode:  "disable",
	}

	db := New(cfg)
	assert.NotNil(t, db)
	assert.Equal(t, "postgres", db.User)
	assert.Equal(t, 5432, db.Port)
	assert.Nil(t, db.Pool)
}
