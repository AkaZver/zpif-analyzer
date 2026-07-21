package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "zpif", cfg.DBUser)
	assert.Equal(t, "zpif", cfg.DBPassword)
	assert.Equal(t, "zpif_analyzer", cfg.DBName)
	assert.Equal(t, "disable", cfg.DBSSLMode)
	assert.Equal(t, "https://api.openai.com/v1", cfg.OpenAIBaseURL)
	assert.Equal(t, "gpt-4o-mini", cfg.OpenAIModel)
	assert.Equal(t, "change-me-in-production", cfg.JWTSecret)
	assert.Equal(t, "8080", cfg.ServerPort)
}

func TestLoad_FromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "myuser")
	os.Setenv("DB_PASSWORD", "mypass")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("DB_SSL_MODE", "require")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("OPENAI_BASE_URL", "https://custom.api.com")
	os.Setenv("OPENAI_MODEL", "gpt-4")
	os.Setenv("JWT_SECRET", "my-secret")
	os.Setenv("SERVER_PORT", "9090")

	cfg := Load()

	assert.Equal(t, "myhost", cfg.DBHost)
	assert.Equal(t, 5433, cfg.DBPort)
	assert.Equal(t, "myuser", cfg.DBUser)
	assert.Equal(t, "mypass", cfg.DBPassword)
	assert.Equal(t, "mydb", cfg.DBName)
	assert.Equal(t, "require", cfg.DBSSLMode)
	assert.Equal(t, "test-key", cfg.OpenAIAPIKey)
	assert.Equal(t, "https://custom.api.com", cfg.OpenAIBaseURL)
	assert.Equal(t, "gpt-4", cfg.OpenAIModel)
	assert.Equal(t, "my-secret", cfg.JWTSecret)
	assert.Equal(t, "9090", cfg.ServerPort)
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_PORT", "not-a-number")

	cfg := Load()

	assert.Equal(t, 5432, cfg.DBPort)
}

func TestGetEnv(t *testing.T) {
	os.Clearenv()

	assert.Equal(t, "default", getEnv("NONEXISTENT", "default"))

	os.Setenv("TEST_KEY", "test_value")
	assert.Equal(t, "test_value", getEnv("TEST_KEY", "default"))
}

func TestGetEnvInt(t *testing.T) {
	os.Clearenv()

	assert.Equal(t, 42, getEnvInt("NONEXISTENT", 42))

	os.Setenv("TEST_INT", "100")
	assert.Equal(t, 100, getEnvInt("TEST_INT", 42))

	os.Setenv("TEST_INVALID", "abc")
	assert.Equal(t, 42, getEnvInt("TEST_INVALID", 42))
}
