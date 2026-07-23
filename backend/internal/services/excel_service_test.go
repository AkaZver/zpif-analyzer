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
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}))

	data, err := service.ExportToExcel()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 0)
}

func TestExcelService_ExportToExcel_WithFunds(t *testing.T) {
	service, mock, cleanup := setupTestExcelService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "PARUS", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(rows)

	// Mock preloads
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	data, err := service.ExportToExcel()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 0)

	// Verify we can open the exported file
	f, err := excelize.OpenReader(bytes.NewReader(data))
	assert.NoError(t, err)
	if f != nil {
		defer f.Close()

		// Check that sheets exist
		idx1, _ := f.GetSheetIndex("Фонды")
		assert.NotEqual(t, -1, idx1)
		idx2, _ := f.GetSheetIndex("Финансы")
		assert.NotEqual(t, -1, idx2)
		idx3, _ := f.GetSheetIndex("Анализ")
		assert.NotEqual(t, -1, idx3)
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
