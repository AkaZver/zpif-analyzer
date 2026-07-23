package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"github.com/zpif-analyzer/backend/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestFundHandler(t *testing.T) (*FundHandler, *gin.Engine, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	fundRepo := repositories.NewFundRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)

	fundService := services.NewFundService(fundRepo, financialsRepo, documentRepo, analysisRepo)
	handler := NewFundHandler(fundService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return handler, router, mock, cleanup
}

func TestFundHandler_GetAllFunds(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(	[]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "").
		AddRow(2, now, now, nil, "Акцент 5", "RU000A10DQF7", "", "Акцент", "офисы", true, false, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	router.GET("/api/funds", handler.GetAllFunds)

	req := httptest.NewRequest("GET", "/api/funds", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var funds []models.Fund
	err := json.Unmarshal(w.Body.Bytes(), &funds)
	assert.NoError(t, err)
	assert.Len(t, funds, 2)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
}

func TestFundHandler_GetFundByID(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(	[]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(1, 1).
		WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	router.GET("/api/funds/:id", handler.GetFundByID)

	req := httptest.NewRequest("GET", "/api/funds/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fund models.Fund
	err := json.Unmarshal(w.Body.Bytes(), &fund)
	assert.NoError(t, err)
	assert.Equal(t, "Парус ОЗН", fund.Name)
	assert.Equal(t, "PARUS", fund.Ticker)
}

func TestFundHandler_GetFundByID_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.GET("/api/funds/:id", handler.GetFundByID)

	req := httptest.NewRequest("GET", "/api/funds/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetFundByID_NotFound(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.GET("/api/funds/:id", handler.GetFundByID)

	req := httptest.NewRequest("GET", "/api/funds/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFundHandler_CreateFund(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(`isin = .+ AND "funds"\."deleted_at" IS NULL`).
		WithArgs("RU000NEW001", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit()

	router.POST("/api/funds", handler.CreateFund)

	body := map[string]interface{}{
		"name":               "Новый фонд",
		"isin":               "RU000NEW001",
		"management_company": "Тест УК",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/funds", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFundHandler_CreateFund_EmptyISIN(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.POST("/api/funds", handler.CreateFund)

	body := map[string]interface{}{
		"name": "Fund without ISIN",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/funds", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_DeleteFund(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_financials" SET "deleted_at"=`).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_documents" SET "deleted_at"=`).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "llm_analyses" SET "deleted_at"=`).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "funds" WHERE "funds"\."id" =`).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	router.DELETE("/api/funds/:id", handler.DeleteFund)

	req := httptest.NewRequest("DELETE", "/api/funds/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFundHandler_DeleteFund_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.DELETE("/api/funds/:id", handler.DeleteFund)

	req := httptest.NewRequest("DELETE", "/api/funds/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetDiscoveryStatus(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.GET("/api/funds/:id/discovery-status", handler.GetDiscoveryStatus)

	req := httptest.NewRequest("GET", "/api/funds/1/discovery-status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var status map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &status)
	assert.NoError(t, err)
	assert.Equal(t, "idle", status["status"])
}

func TestFundHandler_GetDiscoveryStatus_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.GET("/api/funds/:id/discovery-status", handler.GetDiscoveryStatus)

	req := httptest.NewRequest("GET", "/api/funds/abc/discovery-status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_DiscoverDocuments(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.POST("/api/funds/:id/discover", handler.DiscoverDocuments)

	req := httptest.NewRequest("POST", "/api/funds/1/discover", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFundHandler_DiscoverAllDocuments(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.POST("/api/funds/discover-all", handler.DiscoverAllDocuments)

	req := httptest.NewRequest("POST", "/api/funds/discover-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFundHandler_UpdateFund(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows(	[]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Old Name", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	router.PUT("/api/funds/:id", handler.UpdateFund)

	body := map[string]interface{}{
		"name":               "Updated Name",
		"isin":               "RU000A1022Z1",
		"management_company": "Парус",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/funds/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFundHandler_UpdateFund_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.PUT("/api/funds/:id", handler.UpdateFund)

	req := httptest.NewRequest("PUT", "/api/funds/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetFinancialsByFundID(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "snapshot_date",
		"unit_price_rub", "nav_per_unit_rub", "nav_total_mln_rub", "discount_to_nav_pct",
		"cap_rate_pct", "p_nav", "p_affo", "noi_yield_pct",
		"annual_payout_rub", "payout_yield_pct", "payout_yield_after_tax_pct",
		"payout_frequency", "payout_stability", "rent_indexation_pct",
		"management_fee_pct", "trading_volume_mln_rub",
		"number_of_properties", "main_tenants",
	}).AddRow(1, now, now, nil, 1, snapshotDate, 1000.0, 1050.0, 5000.0, -4.76,
		8.5, 0.95, 12.0, 7.2,
		80.0, 8.0, 6.96,
		"monthly", "high", 3.0,
		1.5, 5.0,
		3, "Ozon")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_financials" WHERE fund_id = $1 AND "fund_financials"."deleted_at" IS NULL ORDER BY snapshot_date DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(rows)

	router.GET("/api/funds/:id/financials", handler.GetFinancialsByFundID)

	req := httptest.NewRequest("GET", "/api/funds/1/financials", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFundHandler_GetFinancialsByFundID_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.GET("/api/funds/:id/financials", handler.GetFinancialsByFundID)

	req := httptest.NewRequest("GET", "/api/funds/abc/financials", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_AddFinancials(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус", "RU000A1022Z1", "", "Парус", "", false, false, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	router.POST("/api/funds/:id/financials", handler.AddFinancials)

	body := map[string]interface{}{
		"snapshot_date":    "2024-01-15T00:00:00Z",
		"unit_price_rub":   1000.0,
		"nav_per_unit_rub": 1050.0,
		"payout_frequency": "monthly",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/funds/1/financials", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFundHandler_AddFinancials_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.POST("/api/funds/:id/financials", handler.AddFinancials)

	req := httptest.NewRequest("POST", "/api/funds/abc/financials", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetDocumentsByFundID(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "/docs/report.pdf", "appraisal",
		"abc123", "auto", "https://example.com", now, "downloaded")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY upload_date DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(rows)

	router.GET("/api/funds/:id/documents", handler.GetDocumentsByFundID)

	req := httptest.NewRequest("GET", "/api/funds/1/documents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFundHandler_GetDocumentsByFundID_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.GET("/api/funds/:id/documents", handler.GetDocumentsByFundID)

	req := httptest.NewRequest("GET", "/api/funds/abc/documents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetLatestAnalysis(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "document_id", "model_used", "raw_response",
		"analysis_summary", "risk_assessment", "pros_cons", "extracted_metrics",
	}).AddRow(1, now, now, nil, 1, 1, "gpt-4", "raw", "summary", "low risk", "pros: good", "{}")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY created_at DESC,"llm_analyses"."id" LIMIT $2`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	router.GET("/api/funds/:id/analysis", handler.GetLatestAnalysis)

	req := httptest.NewRequest("GET", "/api/funds/1/analysis", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFundHandler_GetLatestAnalysis_InvalidID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.GET("/api/funds/:id/analysis", handler.GetLatestAnalysis)

	req := httptest.NewRequest("GET", "/api/funds/abc/analysis", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetLatestAnalysis_NotFound(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY created_at DESC,"llm_analyses"."id" LIMIT $2`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.GET("/api/funds/:id/analysis", handler.GetLatestAnalysis)

	req := httptest.NewRequest("GET", "/api/funds/999/analysis", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFundHandler_DeleteDocument_InvalidDocID(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.DELETE("/api/funds/documents/:id", handler.DeleteDocument)

	req := httptest.NewRequest("DELETE", "/api/funds/documents/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_DeleteDocument_Success(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fund_documents" SET "deleted_at"=$1 WHERE "fund_documents"."id" = $2 AND "fund_documents"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), uint(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	router.DELETE("/api/funds/documents/:id", handler.DeleteDocument)

	req := httptest.NewRequest("DELETE", "/api/funds/documents/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFundHandler_CreateFund_InvalidJSON(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.POST("/api/funds", handler.CreateFund)

	req := httptest.NewRequest("POST", "/api/funds", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_UpdateFund_NotFound(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.PUT("/api/funds/:id", handler.UpdateFund)

	body := map[string]interface{}{"name": "Test"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/funds/999", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetFinancialsByFundID_ServiceError(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_financials" WHERE fund_id = $1 AND "fund_financials"."deleted_at" IS NULL ORDER BY snapshot_date DESC`)).
		WithArgs(uint(1)).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/api/funds/:id/financials", handler.GetFinancialsByFundID)

	req := httptest.NewRequest("GET", "/api/funds/1/financials", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFundHandler_GetDocumentsByFundID_ServiceError(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY upload_date DESC`)).
		WithArgs(uint(1)).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/api/funds/:id/documents", handler.GetDocumentsByFundID)

	req := httptest.NewRequest("GET", "/api/funds/1/documents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
