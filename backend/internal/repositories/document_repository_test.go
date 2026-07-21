package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
)

func TestDocumentRepository_GetByFundID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewDocumentRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "/docs/report.pdf", "appraisal",
		"abc123", "auto", "https://example.com", now, "downloaded")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY upload_date DESC`)).
		WithArgs(1).
		WillReturnRows(rows)

	docs, err := repo.GetByFundID(1)

	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "report.pdf", docs[0].FileName)
	assert.Equal(t, "auto", docs[0].Source)
}

func TestDocumentRepository_GetByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewDocumentRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "/docs/report.pdf", "appraisal",
		"abc123", "manual", "", now, "analyzed")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY "fund_documents"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	doc, err := repo.GetByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Equal(t, "report.pdf", doc.FileName)
}

func TestDocumentRepository_GetByHash(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewDocumentRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "/docs/report.pdf", "appraisal",
		"abc123", "auto", "", now, "downloaded")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE content_hash = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY "fund_documents"."id" LIMIT $2`)).
		WithArgs("abc123", 1).
		WillReturnRows(rows)

	doc, err := repo.GetByHash("abc123")

	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Equal(t, "abc123", doc.ContentHash)
}

func TestDocumentRepository_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewDocumentRepository(gormDB)

	now := time.Now()
	doc := &models.FundDocument{
		FundID:       1,
		FileName:     "new_report.pdf",
		FilePath:     "/docs/new_report.pdf",
		DocumentType: "kid",
		Source:       "manual",
		UploadDate:   now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_documents"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(doc)

	assert.NoError(t, err)
}

func TestDocumentRepository_UpdateStatus(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewDocumentRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fund_documents" SET "status"=$1,"updated_at"=$2 WHERE id = $3 AND "fund_documents"."deleted_at" IS NULL`)).
		WithArgs("analyzed", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(1, "analyzed")

	assert.NoError(t, err)
}

func TestDocumentRepository_Delete(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewDocumentRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fund_documents" SET "deleted_at"=$1 WHERE "fund_documents"."id" = $2 AND "fund_documents"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)

	assert.NoError(t, err)
}
