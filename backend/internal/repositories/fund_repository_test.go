package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

	return gormDB, mock
}

func TestFundRepository_GetAll(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "").
		AddRow(2, now, now, nil, "Акцент 5", "RU000A10DQF7", "", "Акцент", "офисы", true, false, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).WillReturnRows(rows)

	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" IN`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" IN`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" IN`).WillReturnRows(emptyRows)

	funds, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Len(t, funds, 2)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
	assert.Equal(t, "Акцент 5", funds[1].Name)
}

func TestFundRepository_GetByID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" =`).WillReturnRows(emptyRows)

	fund, err := repo.GetByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Equal(t, "Парус ОЗН", fund.Name)
	assert.Equal(t, "PARUS", fund.Ticker)
}

func TestFundRepository_GetByID_NotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."id" =`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund, err := repo.GetByID(999)

	assert.Error(t, err)
	assert.Nil(t, fund)
}

func TestFundRepository_GetByISIN(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(`isin = .+ AND "funds"\."deleted_at" IS NULL`).
		WithArgs("RU000A1022Z1", 1).
		WillReturnRows(rows)

	fund, err := repo.GetByISIN("RU000A1022Z1")

	assert.NoError(t, err)
	assert.NotNil(t, fund)
	assert.Equal(t, "RU000A1022Z1", fund.ISIN)
}

func TestFundRepository_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	fund := &models.Fund{
		Name:              "Тестовый фонд",
		ISIN:              "RU000TEST001",
		ManagementCompany: "Тест УК",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(fund)

	assert.NoError(t, err)
}

func TestFundRepository_Delete(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

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

	err := repo.Delete(1)

	assert.NoError(t, err)
}

func TestFundRepository_Update(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	now := time.Now()
	fund := &models.Fund{
		ID:                1,
		Name:              "Updated Name",
		ISIN:              "RU000A1022Z1",
		ManagementCompany: "Парус",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(fund)

	assert.NoError(t, err)
}

func TestFundRepository_GetAll_Empty(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}))

	funds, err := repo.GetAll()

	assert.NoError(t, err)
	assert.Empty(t, funds)
}

func TestFundRepository_GetAllWithLatestFinancials_Success(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()
	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "").
		AddRow(2, now, now, nil, "Акцент 5", "RU000A10DQF7", "", "Акцент", "офисы", true, false, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).WillReturnRows(fundRows)

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

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(id IN \(SELECT DISTINCT ON \(fund_id\) id FROM fund_financials WHERE fund_id IN`).
		WillReturnRows(financialRows)

	funds, err := repo.GetAllWithLatestFinancials()

	assert.NoError(t, err)
	assert.Len(t, funds, 2)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
	assert.Equal(t, "Акцент 5", funds[1].Name)
}

func TestFundRepository_GetAllWithLatestFinancials_Empty(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}))

	funds, err := repo.GetAllWithLatestFinancials()

	assert.NoError(t, err)
	assert.Empty(t, funds)
}

func TestFundRepository_GetAllWithLatestFinancials_NoFinancials(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFundRepository(gormDB)
	now := time.Now()

	fundRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).WillReturnRows(fundRows)

	emptyFinancialRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "snapshot_date",
		"unit_price_rub", "nav_per_unit_rub", "nav_total_mln_rub", "discount_to_nav_pct",
		"cap_rate_pct", "p_nav", "p_affo", "noi_yield_pct",
		"annual_payout_rub", "payout_yield_pct", "payout_yield_after_tax_pct",
		"payout_frequency", "payout_stability", "rent_indexation_pct",
		"management_fee_pct", "trading_volume_mln_rub",
		"number_of_properties", "main_tenants",
	})

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(id IN \(SELECT DISTINCT ON \(fund_id\) id FROM fund_financials WHERE fund_id IN`).
		WillReturnRows(emptyFinancialRows)

	funds, err := repo.GetAllWithLatestFinancials()

	assert.NoError(t, err)
	assert.Len(t, funds, 1)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
	assert.Empty(t, funds[0].Financials)
}
