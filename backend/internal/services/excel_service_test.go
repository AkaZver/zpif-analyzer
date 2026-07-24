package services

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestExcelService(t *testing.T) (*ExcelService, sqlmock.Sqlmock, func()) {
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

	fundRepo := repositories.NewFundRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	service := NewExcelService(fundRepo, financialsRepo, analysisRepo)

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return service, mock, cleanup
}

func TestExcelService_ExportToExcel_Empty(t *testing.T) {
	service, mock, cleanup := setupTestExcelService(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}))

	data, err := service.ExportToExcel()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 0)
}

func TestExcelService_ExportToExcel_WithFunds(t *testing.T) {
	service, mock, cleanup := setupTestExcelService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "investfunds_url", "vsezpif_url"}).
		AddRow(1, now, now, nil, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(rows)

	// Mock preloads - use generic pattern for all three preloads
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" = \$1`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" = \$1`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" = \$1`).WillReturnRows(emptyRows)

	// Mock financials query in ExportToExcel
	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE fund_id = \$1`).WillReturnRows(emptyRows)

	data, err := service.ExportToExcel()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 0)

	// Verify we can open the exported file
	f, err := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, err)
	if f != nil {
		defer f.Close()

		// Check that only "Фонды" sheet exists
		idx1, _ := f.GetSheetIndex("Фонды")
		assert.NotEqual(t, -1, idx1)

		// Check headers
		headers := []string{"Название", "ISIN", "Тикер", "Квал", "УК", "Сегмент",
			"Цена пая", "РСП", "Дисконт к РСП", "Cap Rate", "СЧА, млн ₽",
			"P/NAV", "P/AFFO", "Доходность выплат", "Комиссия УК"}
		for i, expected := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			val, _ := f.GetCellValue("Фонды", cell)
			assert.Equal(t, expected, val)
		}

		// Check fund data in row 2
		name, _ := f.GetCellValue("Фонды", "A2")
		assert.Equal(t, "Парус ОЗН", name)
		isin, _ := f.GetCellValue("Фонды", "B2")
		assert.Equal(t, "RU000A1022Z1", isin)
	}
}

func TestBoolToString(t *testing.T) {
	assert.Equal(t, "Да", boolToString(true))
	assert.Equal(t, "Нет", boolToString(false))
}

func TestCellName(t *testing.T) {
	assert.Equal(t, "A1", cellName(1, 1))
	assert.Equal(t, "B2", cellName(2, 2))
	assert.Equal(t, "Z10", cellName(26, 10))
}
