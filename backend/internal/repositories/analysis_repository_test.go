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

func TestAnalysisRepository_GetLatestByFundID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewAnalysisRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "document_id", "model_used", "raw_response",
		"analysis_summary", "risk_assessment", "pros_cons", "extracted_metrics",
	}).AddRow(1, now, now, nil, 1, 1, "gpt-4", "raw", "summary", "low risk", "pros: good", "{}")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY created_at DESC,"llm_analyses"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	analysis, err := repo.GetLatestByFundID(1)

	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, uint(1), analysis.FundID)
	assert.Equal(t, "gpt-4", analysis.ModelUsed)
	assert.Equal(t, "summary", analysis.AnalysisSummary)
}

func TestAnalysisRepository_GetLatestByFundID_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewAnalysisRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY created_at DESC,"llm_analyses"."id" LIMIT $2`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	analysis, err := repo.GetLatestByFundID(999)

	assert.Error(t, err)
	assert.Nil(t, analysis)
}

func TestAnalysisRepository_GetByFundID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewAnalysisRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "document_id", "model_used", "raw_response",
		"analysis_summary", "risk_assessment", "pros_cons", "extracted_metrics",
	}).AddRow(1, now, now, nil, 1, 1, "gpt-4", "raw1", "summary1", "low", "pros1", "{}").
		AddRow(2, now, now, nil, 1, 2, "gpt-4", "raw2", "summary2", "medium", "pros2", "{}")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY created_at DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(rows)

	analyses, err := repo.GetByFundID(1)

	assert.NoError(t, err)
	assert.Len(t, analyses, 2)
	assert.Equal(t, "summary1", analyses[0].AnalysisSummary)
}

func TestAnalysisRepository_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewAnalysisRepository(gormDB)

	analysis := &models.LLMAnalysis{
		FundID:          1,
		ModelUsed:       "gpt-4",
		AnalysisSummary: "test summary",
		RawResponse:     "raw response",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_analyses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(analysis)

	assert.NoError(t, err)
}

func TestAnalysisRepository_GetByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewAnalysisRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "document_id", "model_used", "raw_response",
		"analysis_summary", "risk_assessment", "pros_cons", "extracted_metrics",
	}).AddRow(1, now, now, nil, 1, 1, "gpt-4", "raw", "summary", "low", "pros", "{}")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE "llm_analyses"."id" = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY "llm_analyses"."id" LIMIT $2`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	analysis, err := repo.GetByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "gpt-4", analysis.ModelUsed)
}

func TestAnalysisRepository_GetByID_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewAnalysisRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE "llm_analyses"."id" = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY "llm_analyses"."id" LIMIT $2`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	analysis, err := repo.GetByID(999)

	assert.Error(t, err)
	assert.Nil(t, analysis)
}
