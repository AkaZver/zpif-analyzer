package services

import (
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

type ExcelService struct {
	fundRepo       *repositories.FundRepository
	financialsRepo *repositories.FinancialsRepository
}

func NewExcelService(
	fundRepo *repositories.FundRepository,
	financialsRepo *repositories.FinancialsRepository,
) *ExcelService {
	return &ExcelService{
		fundRepo:       fundRepo,
		financialsRepo: financialsRepo,
	}
}

func (s *ExcelService) ExportToExcel() ([]byte, error) {
	// TODO: Implement Excel export logic
	// This will be implemented in Phase 4
	return []byte{}, nil
}

func (s *ExcelService) ImportFromExcel(data []byte) (int, error) {
	// TODO: Implement Excel import logic
	// This will be implemented in Phase 4
	return 0, nil
}

// Helper function to get all funds with financials for export
func (s *ExcelService) GetAllFundsData() ([]models.Fund, error) {
	return s.fundRepo.GetAll()
}

// Helper function to import funds
func (s *ExcelService) ImportFund(fund *models.Fund) error {
	return s.fundRepo.Create(fund)
}

// Helper function to import financials
func (s *ExcelService) ImportFinancials(financials *models.FundFinancials) error {
	return s.financialsRepo.Create(financials)
}
