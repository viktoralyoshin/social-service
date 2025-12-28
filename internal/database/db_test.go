package database

import (
	"social-service/internal/config"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInit_Error(t *testing.T) {
	cfg := &config.Config{
		DBHost: "invalid_host",
	}

	db, err := Init(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestMigrate_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		mock.ExpectClose()
		_ = db.Close()
	}()

	err = Migrate(db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrations directory does not exist")

	assert.NoError(t, mock.ExpectationsWereMet())
}
