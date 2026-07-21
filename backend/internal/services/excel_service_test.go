package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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
	service := NewExcelService(fundRepo, financialsRepo)

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return service, mock, cleanup
}

func TestExcelService_ExportToExcel(t *testing.T) {
	service, _, cleanup := setupTestExcelService(t)
	defer cleanup()

	data, err := service.ExportToExcel()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	// Currently returns empty byte slice (TODO)
}

func TestExcelService_ImportFromExcel(t *testing.T) {
	service, _, cleanup := setupTestExcelService(t)
	defer cleanup()

	imported, err := service.ImportFromExcel([]byte{})

	assert.NoError(t, err)
	assert.Equal(t, 0, imported)
	// Currently returns 0 (TODO)
}
