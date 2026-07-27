package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zpif-analyzer/backend/internal/llm"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MockAnalyzer struct {
	mock.Mock
}

func (m *MockAnalyzer) AnalyzeLatestDocuments(ctx context.Context, fund *models.Fund) (*models.LLMAnalysis, error) {
	args := m.Called(ctx, fund)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LLMAnalysis), args.Error(1)
}

func (m *MockAnalyzer) AnalyzeDocuments(ctx context.Context, fund *models.Fund, documentIDs []uint) (*models.LLMAnalysis, error) {
	args := m.Called(ctx, fund, documentIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LLMAnalysis), args.Error(1)
}

type MockDiscoverer struct {
	mock.Mock
}

func (m *MockDiscoverer) Discover(ctx context.Context, fund *models.Fund) (*llm.DiscoveryStatus, error) {
	args := m.Called(ctx, fund)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.DiscoveryStatus), args.Error(1)
}

func (m *MockDiscoverer) GetStatus(fundID uint) *llm.DiscoveryStatus {
	args := m.Called(fundID)
	return args.Get(0).(*llm.DiscoveryStatus)
}

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
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds"`)).WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	funds, err := service.GetAllFunds()

	assert.NoError(t, err)
	assert.Len(t, funds, 1)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
}

func TestFundService_GetAllFundsWithLatestFinancials_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "").
		AddRow(2, now, now, nil, "Акцент 5", "RU000A10DQF7", "", "Акцент", "офисы", true, false, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds"`)).WillReturnRows(fundRows)

	financialRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "snapshot_date",
		"unit_price_rub", "nav_per_unit_rub", "nav_total_mln_rub", "discount_to_nav_pct",
		"cap_rate_pct", "p_nav", "p_affo", "noi_yield_pct",
		"annual_payout_rub", "payout_yield_pct", "payout_yield_after_tax_pct",
		"payout_frequency", "payout_stability", "rent_indexation_pct",
		"management_fee_pct", "trading_volume_mln_rub",
		"number_of_properties", "main_tenants",
	}).
		AddRow(1, now, now, nil, 1, snapshotDate, 1000.0, 1050.0, 5000.0, -4.76,
			8.5, 0.95, 12.0, 7.2, 80.0, 8.0, 6.96, "monthly", "high", 3.0, 1.5, 5.0, 3, "Ozon").
		AddRow(2, now, now, nil, 2, snapshotDate, 900.0, 950.0, 4000.0, -5.26,
			7.5, 0.90, 11.0, 6.5, 70.0, 7.0, 6.09, "quarterly", "medium", 2.5, 1.2, 4.0, 2, "Wildberries")

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE id IN \(SELECT DISTINCT ON \(fund_id\) id FROM fund_financials WHERE fund_id IN`).
		WillReturnRows(financialRows)

	funds, err := service.GetAllFundsWithLatestFinancials()

	assert.NoError(t, err)
	assert.Len(t, funds, 2)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
	assert.Equal(t, "Акцент 5", funds[1].Name)
}

func TestFundService_GetAllFundsWithLatestFinancials_Empty(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}))

	funds, err := service.GetAllFundsWithLatestFinancials()

	assert.NoError(t, err)
	assert.Empty(t, funds)
}

func TestFundService_GetFundByID(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, "", "")

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
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Existing", "RU000A1022Z1", "", "Парус", "", false, false, nil, "", "")

	mock.ExpectQuery(`isin = .+`).
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

	mock.ExpectQuery(`isin = .+`).
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
	mock.ExpectExec(`DELETE FROM "fund_financials" WHERE fund_id =`).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "fund_documents" WHERE fund_id =`).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "llm_analyses" WHERE fund_id =`).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "funds" WHERE "funds"\."id" =`).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteFund(1)

	assert.NoError(t, err)
}

func TestFundService_DiscoverDocumentsForFund(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	err := service.DiscoverDocumentsForFund(context.Background(), 1)

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
	mock.ExpectExec(`DELETE FROM "fund_documents" WHERE "fund_documents"\."id" =`).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := service.DeleteDocument(1)

	assert.NoError(t, err)
}

func TestFundService_UpdateFund_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Old Name", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "")

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

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE fund_id = .+ AND snapshot_date <= .+ ORDER BY snapshot_date DESC`).
		WithArgs(uint(1), sqlmock.AnyArg()).
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

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE fund_id = .+ AND snapshot_date <= .+ ORDER BY snapshot_date DESC,"fund_financials"\."id" LIMIT`).
		WithArgs(uint(1), sqlmock.AnyArg(), 1).
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 ORDER BY upload_date DESC`)).
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
	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус", "RU000A1022Z1", "", "Парус", "", false, false, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE content_hash = $1 ORDER BY "fund_documents"."id" LIMIT $2`)).
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
	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус", "RU000A1022Z1", "", "Парус", "", false, false, nil, "", "")

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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE content_hash = $1 ORDER BY "fund_documents"."id" LIMIT $2`)).
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_analyses" WHERE fund_id = $1 ORDER BY created_at DESC,"llm_analyses"."id" LIMIT $2`)).
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

func TestFundService_AnalyzeFund_AnalyzerNotConfigured(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	// Analyzer is nil by default in setupTestService
	result, err := service.AnalyzeFund(context.Background(), 1, []uint{1, 2, 3})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "analyzer not configured")
}

func TestFundService_AnalyzeFund_FundNotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	// Mock analyzer
	analyzer := &MockAnalyzer{}
	service.SetAnalyzer(analyzer)

	// Mock fund not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := service.AnalyzeFund(context.Background(), 999, []uint{1, 2, 3})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFundService_AnalyzeFund_WithDocumentIDs(t *testing.T) {
	service, mockDB, cleanup := setupTestService(t)
	defer cleanup()

	// Mock analyzer
	analyzer := &MockAnalyzer{}
	analyzer.On("AnalyzeDocuments", mock.Anything, mock.Anything, []uint{1, 2, 3}).
		Return(&models.LLMAnalysis{ID: 1, FundID: 1, ModelUsed: "gpt-4"}, nil)
	service.SetAnalyzer(analyzer)

	// Mock fund exists
	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Test Fund", "RU000TEST001", "", "Test UK", "", false, false, nil, "", "")
	
	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	result, err := service.AnalyzeFund(context.Background(), 1, []uint{1, 2, 3})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "gpt-4", result.ModelUsed)
}

func TestFundService_AnalyzeFund_WithoutDocumentIDs(t *testing.T) {
	service, mockDB, cleanup := setupTestService(t)
	defer cleanup()

	// Mock analyzer
	analyzer := &MockAnalyzer{}
	analyzer.On("AnalyzeLatestDocuments", mock.Anything, mock.Anything).
		Return(&models.LLMAnalysis{ID: 1, FundID: 1, ModelUsed: "gpt-4"}, nil)
	service.SetAnalyzer(analyzer)

	// Mock fund exists
	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Test Fund", "RU000TEST001", "", "Test UK", "", false, false, nil, "", "")
	
	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	result, err := service.AnalyzeFund(context.Background(), 1, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "gpt-4", result.ModelUsed)
}

func TestFundService_GetDocumentByID_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "/docs/report.pdf", "appraisal",
		"abc123", "auto", "", now, "downloaded")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	doc, err := service.GetDocumentByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Equal(t, "report.pdf", doc.FileName)
}

func TestFundService_GetDocumentByID_NotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	doc, err := service.GetDocumentByID(999)

	assert.Error(t, err)
	assert.Nil(t, doc)
}

func TestFundService_GetDiscoveryStatus_NilDiscoverer(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	status := service.GetDiscoveryStatus(1)

	assert.Equal(t, "idle", status["status"])
	assert.Equal(t, uint(1), status["fund_id"])
}

func TestFundService_GetDiscoveryStatus_WithDiscoverer(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	discoverer := &MockDiscoverer{}
	discoverer.On("GetStatus", uint(1)).Return(&llm.DiscoveryStatus{
		FundID: 1,
		Status: "completed",
	})
	service.SetDiscoverer(discoverer)

	status := service.GetDiscoveryStatus(1)

	assert.Equal(t, "completed", status["status"])
	assert.Equal(t, uint(1), status["fund_id"])
}

func TestFundService_SetDiscoverer(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	assert.Nil(t, service.discoverer)

	discoverer := &MockDiscoverer{}
	service.SetDiscoverer(discoverer)

	assert.NotNil(t, service.discoverer)
}

func TestFundService_SetAnalyzer(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	assert.Nil(t, service.analyzer)

	analyzer := &MockAnalyzer{}
	service.SetAnalyzer(analyzer)

	assert.NotNil(t, service.analyzer)
}

func TestFundService_SetLLMSettingsRepo(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	assert.Nil(t, service.llmSettingsRepo)

	db, _, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	repo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(repo)

	assert.NotNil(t, service.llmSettingsRepo)
}

func TestFundService_DiscoverDocumentsForFund_WithDiscoverer_FundNotFound(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	discoverer := &MockDiscoverer{}
	service.SetDiscoverer(discoverer)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	err := service.DiscoverDocumentsForFund(context.Background(), 999)

	assert.Error(t, err)
}

func TestFundService_DiscoverDocumentsForFund_WithDiscoverer_Success(t *testing.T) {
	service, mockDB, cleanup := setupTestService(t)
	defer cleanup()

	discoverer := &MockDiscoverer{}
	discoverer.On("Discover", mock.Anything, mock.Anything).Return(&llm.DiscoveryStatus{
		FundID: 1,
		Status: "completed",
	}, nil)
	service.SetDiscoverer(discoverer)

	now := time.Now()
	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Test Fund", "RU000TEST001", "", "Test UK", "", false, false, nil, "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(fundRows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	err := service.DiscoverDocumentsForFund(context.Background(), 1)

	assert.NoError(t, err)
}

func TestFundService_DiscoverDocumentsForAllFunds_WithDiscoverer(t *testing.T) {
	service, mockDB, cleanup := setupTestService(t)
	defer cleanup()

	discoverer := &MockDiscoverer{}
	discoverer.On("Discover", mock.Anything, mock.Anything).Return(&llm.DiscoveryStatus{
		FundID: 1,
		Status: "completed",
	}, nil)
	service.SetDiscoverer(discoverer)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Fund 1", "RU000A1022Z1", "", "UK", "", false, false, nil, "", "").
		AddRow(2, now, now, nil, "Fund 2", "RU000A1022Z2", "", "UK", "", false, false, nil, "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds"`)).WillReturnRows(rows)
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mockDB.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	err := service.DiscoverDocumentsForAllFunds()

	assert.NoError(t, err)
}

func TestFundService_EnrichAndCreateFund_LLMSettingsNil(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	fund, err := service.EnrichAndCreateFund(context.Background(), "Парус ОЗН")

	assert.Error(t, err)
	assert.Nil(t, fund)
	assert.Contains(t, err.Error(), "LLM settings not configured")
}

func TestFundService_EnrichAndCreateFund_SettingsError(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund, err := service.EnrichAndCreateFund(context.Background(), "Парус ОЗН")

	assert.Error(t, err)
	assert.Nil(t, fund)
	assert.Contains(t, err.Error(), "failed to load LLM settings")
}

func TestFundService_EnrichAndCreateFund_APIKeyEmpty(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "", "http://localhost", "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	fund, err := service.EnrichAndCreateFund(context.Background(), "Парус ОЗН")

	assert.Error(t, err)
	assert.Nil(t, fund)
	assert.Contains(t, err.Error(), "LLM API key not configured")
}

func TestFundService_EnrichAndCreateFund_LLMCallFails(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	fund, err := service.EnrichAndCreateFund(context.Background(), "Парус ОЗН")

	assert.Error(t, err)
	assert.Nil(t, fund)
	assert.Contains(t, err.Error(), "LLM call failed")
}

func TestFundService_EnrichAndCreateFund_Success(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": `{"name":"ЗПИФ Парус","isin":"RU000A1022Z1","ticker":"PARUS","management_company":"Парус УК","real_estate_segment":"склады","qualified_required":false,"has_market_maker":true,"fund_end_date":null}`,
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	mock.ExpectQuery(`isin = .+`).
		WithArgs("RU000A1022Z1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit()

	fund, err := service.EnrichAndCreateFund(context.Background(), "Парус ОЗН")

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Equal(t, "ЗПИФ Парус", fund.Name)
	assert.Equal(t, "RU000A1022Z1", fund.ISIN)
}

func TestFundService_EnrichAndCreateFund_PendingISIN(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": `{"name":"Неизвестный фонд","isin":"UNKNOWN","ticker":"","management_company":"УК","real_estate_segment":"","qualified_required":false,"has_market_maker":false,"fund_end_date":null}`,
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	mock.ExpectQuery(`isin = .+`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))
	mock.ExpectCommit()

	fund, err := service.EnrichAndCreateFund(context.Background(), "Неизвестный фонд")

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Contains(t, fund.ISIN, "PENDING-")
}

func TestFundService_EnrichAndCreateFund_WithFundEndDate(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": `{"name":"Фонд с датой","isin":"RU000A102N77","ticker":"","management_company":"УК","real_estate_segment":"офисы","qualified_required":true,"has_market_maker":false,"fund_end_date":"2030-12-31"}`,
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	mock.ExpectQuery(`isin = .+`).
		WithArgs("RU000A102N77", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	fund, err := service.EnrichAndCreateFund(context.Background(), "Фонд с датой")

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.NotNil(t, fund.FundEndDate)
	assert.Equal(t, 2030, fund.FundEndDate.Year())
	assert.Equal(t, time.December, fund.FundEndDate.Month())
}

func TestFundService_EnrichAndCreateFund_InvalidJSON(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "not valid json at all",
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	fund, err := service.EnrichAndCreateFund(context.Background(), "Тест")

	assert.Error(t, err)
	assert.Nil(t, fund)
	assert.Contains(t, err.Error(), "failed to parse LLM response")
}

func TestFundService_EnrichAndCreateFund_CreateFundFails(t *testing.T) {
	service, mock, cleanup := setupTestService(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": `{"name":"Дубликат","isin":"RU000A1022Z1","ticker":"","management_company":"УК","real_estate_segment":"","qualified_required":false,"has_market_maker":false,"fund_end_date":null}`,
					},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	db, mockDB, _ := sqlmock.New()
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	service.SetLLMSettingsRepo(settingsRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mockDB.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	existingFundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Existing", "RU000A1022Z1", "", "UK", "", false, false, nil, "", "")

	mock.ExpectQuery(`isin = .+`).
		WithArgs("RU000A1022Z1", 1).
		WillReturnRows(existingFundRows)

	fund, err := service.EnrichAndCreateFund(context.Background(), "Дубликат")

	assert.Error(t, err)
	assert.Nil(t, fund)
	assert.Contains(t, err.Error(), "already exists")
}
