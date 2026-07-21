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
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}))

	data, err := service.ExportToExcel()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 0)
}

func TestExcelService_ExportToExcel_WithFunds(t *testing.T) {
	service, mock, cleanup := setupTestExcelService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
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

func TestExcelService_ImportFromExcel_InvalidData(t *testing.T) {
	service, _, cleanup := setupTestExcelService(t)
	defer cleanup()

	imported, err := service.ImportFromExcel([]byte("invalid data"))

	assert.Error(t, err)
	assert.Equal(t, 0, imported)
	assert.Contains(t, err.Error(), "failed to open excel")
}

func TestExcelService_ImportFromExcel_EmptyFile(t *testing.T) {
	service, _, cleanup := setupTestExcelService(t)
	defer cleanup()

	// Create empty Excel file
	f := excelize.NewFile()
	var buf bytes.Buffer
	f.Write(&buf)

	imported, err := service.ImportFromExcel(buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 0, imported)
}

func TestExcelService_ImportFromExcel_WithFunds(t *testing.T) {
	service, mock, cleanup := setupTestExcelService(t)
	defer cleanup()

	// Create Excel file with test data
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Фонды")
	
	headers := []string{"ID", "Название", "ISIN", "Тикер", "УК", "Сегмент", "Квал", "ММ"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Фонды", cell, h)
	}
	
	f.SetCellValue("Фонды", "A2", 1)
	f.SetCellValue("Фонды", "B2", "Тестовый фонд")
	f.SetCellValue("Фонды", "C2", "RU000TEST001")
	f.SetCellValue("Фонды", "D2", "TEST")
	f.SetCellValue("Фонды", "E2", "Тест УК")
	f.SetCellValue("Фонды", "F2", "офисы")
	f.SetCellValue("Фонды", "G2", "Нет")
	f.SetCellValue("Фонды", "H2", "Да")

	var buf bytes.Buffer
	f.Write(&buf)

	// Mock GetByISIN returns not found
	mock.ExpectQuery(`isin = .+ AND "funds"\."deleted_at" IS NULL`).
		WithArgs("RU000TEST001", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock Create
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "funds"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	imported, err := service.ImportFromExcel(buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 1, imported)
}

func TestExcelService_GetAllFundsData(t *testing.T) {
	service, mock, cleanup := setupTestExcelService(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_start_date", "fund_end_date", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "Парус ОЗН", "RU000A1022Z1", "", "Парус", "склады", false, true, nil, nil, now, now, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(rows)
	
	emptyRows := sqlmock.NewRows([]string{"id", "fund_id"})
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)
	mock.ExpectQuery(`SELECT \* FROM ".+" WHERE ".+"\."fund_id" (IN|=)`).WillReturnRows(emptyRows)

	funds, err := service.GetAllFundsData()

	assert.NoError(t, err)
	assert.Len(t, funds, 1)
	assert.Equal(t, "Парус ОЗН", funds[0].Name)
}

func TestBoolToString(t *testing.T) {
	assert.Equal(t, "Да", boolToString(true))
	assert.Equal(t, "Нет", boolToString(false))
}

func TestStringToBool(t *testing.T) {
	assert.True(t, stringToBool("Да"))
	assert.True(t, stringToBool("да"))
	assert.True(t, stringToBool("Yes"))
	assert.True(t, stringToBool("yes"))
	assert.True(t, stringToBool("true"))
	assert.True(t, stringToBool("1"))
	assert.False(t, stringToBool("Нет"))
	assert.False(t, stringToBool("no"))
	assert.False(t, stringToBool("false"))
	assert.False(t, stringToBool("0"))
}

func TestParseFloat(t *testing.T) {
	assert.Equal(t, 123.45, parseFloat("123.45"))
	assert.Equal(t, 0.0, parseFloat("invalid"))
	assert.Equal(t, 0.0, parseFloat(""))
}

func TestParseInt(t *testing.T) {
	assert.Equal(t, 123, parseInt("123"))
	assert.Equal(t, 0, parseInt("invalid"))
	assert.Equal(t, 0, parseInt(""))
}

func TestCellName(t *testing.T) {
	assert.Equal(t, "A1", cellName(1, 1))
	assert.Equal(t, "B2", cellName(2, 2))
	assert.Equal(t, "Z10", cellName(26, 10))
}
