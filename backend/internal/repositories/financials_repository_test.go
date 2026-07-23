package repositories

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
)

func TestFinancialsRepository_GetByFundID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFinancialsRepository(gormDB)
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
		WithArgs(1).
		WillReturnRows(rows)

	financials, err := repo.GetByFundID(1)

	assert.NoError(t, err)
	assert.Len(t, financials, 1)
	assert.Equal(t, uint(1), financials[0].FundID)
	assert.Equal(t, 1000.0, financials[0].UnitPriceRub)
	assert.Equal(t, "monthly", financials[0].PayoutFrequency)
}

func TestFinancialsRepository_GetLatestByFundID(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFinancialsRepository(gormDB)
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_financials" WHERE fund_id = $1 AND "fund_financials"."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	financial, err := repo.GetLatestByFundID(1)

	assert.NoError(t, err)
	assert.NotNil(t, financial)
	assert.Equal(t, 1000.0, financial.UnitPriceRub)
}

func TestFinancialsRepository_Create(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFinancialsRepository(gormDB)

	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	financial := &models.FundFinancials{
		FundID:          1,
		SnapshotDate:    snapshotDate,
		UnitPriceRub:    1000.0,
		NavPerUnitRub:   1050.0,
		PayoutFrequency: "monthly",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(financial)

	assert.NoError(t, err)
}

func TestFinancialsRepository_Update(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFinancialsRepository(gormDB)

	now := time.Now()
	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	financial := &models.FundFinancials{
		ID:           1,
		FundID:       1,
		SnapshotDate: snapshotDate,
		UnitPriceRub: 1100.0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_financials" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(financial)

	assert.NoError(t, err)
}

func TestFinancialsRepository_Delete(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFinancialsRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fund_financials" SET "deleted_at"=$1 WHERE "fund_financials"."id" = $2 AND "fund_financials"."deleted_at" IS NULL`)).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)

	assert.NoError(t, err)
}

func TestFinancialsRepository_GetByFundIDAndDateRange(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	defer func() { db, _ := gormDB.DB(); db.Close() }()

	repo := NewFinancialsRepository(gormDB)

	now := time.Now()
	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

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

	mock.ExpectQuery(`fund_id = .+ AND snapshot_date BETWEEN`).
		WithArgs(uint(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	financials, err := repo.GetByFundIDAndDateRange(1, from, to)

	assert.NoError(t, err)
	assert.Len(t, financials, 1)
	assert.Equal(t, 1000.0, financials[0].UnitPriceRub)
}
