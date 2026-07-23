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

func setupTestService(t *testing.T) (*FundService, sqlmock.Sqlmock, func()) {
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

	service := NewFundService(fundRepo, financialsRepo, documentRepo, analysisRepo)

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return service, mock, cleanup
}

func TestFundService_GetAllFunds(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	funds, err := service.GetAllFunds()

	assert.NoError(t, err)
	assert.Len(t, funds, 1)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
}

func TestFundService_GetFundByID(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)

	fund, err := service.GetFundByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Equal(t, "Парус ОЗН", fund.Name)
}

func TestFundService_GetFundByID_NotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund, err := service.GetFundByID(999)

	assert.Error(t, err)
	assert.Nil(t, fund)
}

func TestFundService_CreateFund_EmptyISIN(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	fund := &models.Fund{Name: "Test"}

	err := service.CreateFund(fund)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ISIN is required")
}

func TestFundService_CreateFund_DuplicateISIN(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Existing", "RU000A1022Z1", "", "Парус", "", false, false, nil, nil, now, now, nil)

	mock.ExpectQuery(`isin = .+ AND "funds"\."deleted_at" IS NULL`).
		WithArgs("RU000A1022Z1", 1).
		WillReturnRows(rows)

	fund := &models.Fund{Name: "New Fund", ISIN: "RU000A1022Z1"}
	err := service.CreateFund(fund)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestFundService_CreateFund_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(`isin = .+ AND "funds"\."deleted_at" IS NULL`).
		WithArgs("RU000NEW001", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit()

	fund := &models.Fund{Name: "New Fund", ISIN: "RU000NEW001", ManagementCompany: "Test"}
	err := service.CreateFund(fund)

	assert.NoError(t, err)
}

func TestFundService_DeleteFund(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET "deleted_at"=`).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteFund(1)

	assert.NoError(t, err)
}

func TestFundService_DiscoverDocumentsForFund(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	err := service.DiscoverDocumentsForFund(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document discovery not configured")
}

func TestFundService_DiscoverDocumentsForAllFunds(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	err := service.DiscoverDocumentsForAllFunds()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document discovery not configured")
}

func TestFundService_DeleteDocument(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fund_documents" SET "deleted_at"=$1 WHERE "fund_documents"."id" = $2 AND "fund_documents"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteDocument(1)

	assert.NoError(t, err)
}

func TestFundService_UpdateFund_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Old Name", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	fund := &models.Fund{
		Name:              "Updated Name",
		ISIN:              "RU000A1022Z1",
		ManagementCompany: "Парус",
	}
	err := service.UpdateFund(1, fund)

	assert.NoError(t, err)
}

func TestFundService_UpdateFund_NotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund := &models.Fund{Name: "Test"}
	err := service.UpdateFund(999, fund)

	assert.Error(t, err)
}

func TestFundService_GetFinancialsByFundID(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
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
		8.5, 0.95, 12.0, 7.2, 80.0, 8.0, 6.96, "monthly", "high", 3.0, 1.5, 5.0, 3, "Ozon")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_financials" WHERE fund_id = $1 AND "fund_financials"."deleted_at" IS NULL ORDER BY snapshot_date DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(rows)

	financials, err := service.GetFinancialsByFundID(1)

	assert.NoError(t, err)
	assert.Len(t, financials, 1)
	assert.Equal(t, 1000.0, financials[0].UnitPriceRub)
}

func TestFundService_GetLatestFinancials(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
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
		8.5, 0.95, 12.0, 7.2, 80.0, 8.0, 6.96, "monthly", "high", 3.0, 1.5, 5.0, 3, "Ozon")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_financials" WHERE fund_id = $1 AND "fund_financials"."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"."id" LIMIT $2`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	financial, err := service.GetLatestFinancials(1)

	assert.NoError(t, err)
	assert.NotNil(t, financial)
	assert.Equal(t, 1000.0, financial.UnitPriceRub)
}

func TestFundService_AddFinancials_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус", "RU000A1022Z1", "", "Парус", "", false, false, nil, nil, now, now, nil)

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

	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	financials := &models.FundFinancials{
		SnapshotDate:    snapshotDate,
		UnitPriceRub:    1000.0,
		PayoutFrequency: "monthly",
	}
	err := service.AddFinancials(1, financials)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), financials.FundID)
}

func TestFundService_AddFinancials_FundNotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	financials := &models.FundFinancials{}
	err := service.AddFinancials(999, financials)

	assert.Error(t, err)
}

func TestFundService_GetDocumentsByFundID(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "/docs/report.pdf", "appraisal",
		"abc123", "auto", "", now, "downloaded")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY upload_date DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(rows)

	docs, err := service.GetDocumentsByFundID(1)

	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "report.pdf", docs[0].FileName)
}

func TestFundService_AddDocument_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус", "RU000A1022Z1", "", "Парус", "", false, false, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE content_hash = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY "fund_documents"."id" LIMIT $2`)).
		WithArgs("newhash", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_documents"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	doc := &models.FundDocument{
		FundID:      1,
		FileName:    "new_report.pdf",
		ContentHash: "newhash",
	}
	err := service.AddDocument(doc)

	assert.NoError(t, err)
}

func TestFundService_AddDocument_DuplicateHash(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус", "RU000A1022Z1", "", "Парус", "", false, false, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	existingDoc := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "old.pdf", "/docs/old.pdf", "appraisal",
		"existinghash", "auto", "", now, "downloaded")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE content_hash = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY "fund_documents"."id" LIMIT $2`)).
		WithArgs("existinghash", 1).
		WillReturnRows(existingDoc)

	doc := &models.FundDocument{
		FundID:      1,
		ContentHash: "existinghash",
	}
	err := service.AddDocument(doc)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestFundService_AddDocument_FundNotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	doc := &models.FundDocument{FundID: 999}
	err := service.AddDocument(doc)

	assert.Error(t, err)
}

func TestFundService_GetLatestAnalysis(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "document_id", "model_used", "raw_response",
		"analysis_summary", "risk_assessment", "pros_cons", "extracted_metrics",
	}).AddRow(1, now, now, nil, 1, 1, "gpt-4", "raw", "summary", "low risk", "pros", "{}")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 AND "llm_analyses"."deleted_at" IS NULL ORDER BY created_at DESC,"llm_analyses"."id" LIMIT $2`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	analysis, err := service.GetLatestAnalysis(1)

	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "gpt-4", analysis.ModelUsed)
}

func TestFundService_AddAnalysis(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_analyses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	analysis := &models.LLMAnalysis{
		FundID:          1,
		ModelUsed:       "gpt-4",
		AnalysisSummary: "test summary",
	}
	err := service.AddAnalysis(analysis)

	assert.NoError(t, err)
}
