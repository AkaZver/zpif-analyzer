package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

func TestLLMSettingsRepository_Get(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewLLMSettingsRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "model_name",
	}).AddRow(1, now, now, nil, "encrypted_key", "https://api.openai.com/v1", "gpt-4")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	settings, err := repo.Get()

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "https://api.openai.com/v1", settings.BaseURL)
	assert.Equal(t, "gpt-4", settings.ModelName)
}

func TestLLMSettingsRepository_Get_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewLLMSettingsRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnError(gorm.ErrRecordNotFound)

	settings, err := repo.Get()

	assert.Error(t, err)
	assert.Nil(t, settings)
}

func TestLLMSettingsRepository_Upsert_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewLLMSettingsRepository(gormDB)

	settings := &models.LLMSettings{
		APIKeyEncrypted: "key",
		BaseURL:         "https://api.openai.com/v1",
		ModelName:       "gpt-4",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Upsert(settings)

	assert.NoError(t, err)
}

func TestLLMSettingsRepository_Upsert_Update(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewLLMSettingsRepository(gormDB)
	now := time.Now()

	settings := &models.LLMSettings{
		APIKeyEncrypted: "new_key",
		BaseURL:         "https://api.openai.com/v1",
		ModelName:       "gpt-4-turbo",
	}

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "model_name",
	}).AddRow(1, now, now, nil, "old_key", "https://api.openai.com/v1", "gpt-4")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "llm_settings" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Upsert(settings)

	assert.NoError(t, err)
}
