package services

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestLLMService(t *testing.T) (*LLMService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service := NewLLMService(settingsRepo)

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return service, mock, cleanup
}

func TestLLMService_GetSettings_Defaults(t *testing.T) {
	service, mock, cleanup := setupTestLLMService(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "llm_settings"`).
		WillReturnError(gorm.ErrRecordNotFound)

	settings, err := service.GetSettings()

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "https://api.openai.com/v1", settings.BaseURL)
	assert.Equal(t, "gpt-4o-mini", settings.ModelName)
}

func TestLLMService_GetSettings_FromDB(t *testing.T) {
	service, mock, cleanup := setupTestLLMService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "api_key_encrypted", "base_url", "model_name", "created_at", "updated_at",
	}).AddRow(1, "key", "https://custom.api.com", "gpt-4", now, now)

	mock.ExpectQuery(`SELECT \* FROM "llm_settings"`).
		WillReturnRows(rows)

	settings, err := service.GetSettings()

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "https://custom.api.com", settings.BaseURL)
	assert.Equal(t, "gpt-4", settings.ModelName)
}

func TestLLMService_UpdateSettings(t *testing.T) {
	service, mock, cleanup := setupTestLLMService(t)
	defer cleanup()

	// Get returns not found, so Create is called
	mock.ExpectQuery(`SELECT \* FROM "llm_settings"`).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	settings := &models.LLMSettings{
		APIKeyEncrypted: "test_key",
		BaseURL:         "https://api.openai.com/v1",
		ModelName:       "gpt-4",
	}

	err := service.UpdateSettings(settings)

	assert.NoError(t, err)
}

func TestLLMService_TestConnection(t *testing.T) {
	service, mock, cleanup := setupTestLLMService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "model_name",
		"proxy_enabled", "proxy_url", "proxy_username", "proxy_password",
	}).AddRow(1, now, now, nil, "test-api-key", "https://api.openai.com/v1", "gpt-4o-mini",
		false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	err := service.TestConnection()

	assert.Error(t, err)
}

func TestLLMService_UpdateSettings_WithProxy(t *testing.T) {
	service, mock, cleanup := setupTestLLMService(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT \* FROM "llm_settings"`).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	settings := &models.LLMSettings{
		APIKeyEncrypted: "test_key",
		BaseURL:         "https://api.openai.com/v1",
		ModelName:       "gpt-4",
		ProxyEnabled:    true,
		ProxyURL:        "http://proxy.example.com:8080",
		ProxyUsername:   "user",
		ProxyPassword:   "pass",
	}

	err := service.UpdateSettings(settings)

	assert.NoError(t, err)
	assert.True(t, settings.ProxyEnabled)
	assert.Equal(t, "http://proxy.example.com:8080", settings.ProxyURL)
}
