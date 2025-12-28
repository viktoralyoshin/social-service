package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setEnv(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("error while setting env: %v", err)
	}
}

func TestLoad(t *testing.T) {
	originalEnv := os.Getenv("DB_USER")
	defer setEnv(t, "DB_USER", originalEnv)

	setEnv(t, "DB_USER", "test_user")
	setEnv(t, "DB_NAME", "test_db")
	setEnv(t, "DB_HOST", "localhost")
	setEnv(t, "DB_PORT", "5432")
	setEnv(t, "ENV", "local")

	cfg := Load()

	assert.Equal(t, "test_user", cfg.DBUser)
	assert.Equal(t, "test_db", cfg.DBName)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "local", cfg.Env)
}
