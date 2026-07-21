package services

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

type ExcelService struct {
	fundRepo       *repositories.FundRepository
	financialsRepo *repositories.FinancialsRepository
	analysisRepo   *repositories.AnalysisRepository
}

func NewExcelService(
	fundRepo *repositories.FundRepository,
	financialsRepo *repositories.FinancialsRepository,
	analysisRepo *repositories.AnalysisRepository,
) *ExcelService {
	return &ExcelService{
		fundRepo:       fundRepo,
		financialsRepo: financialsRepo,
		analysisRepo:   analysisRepo,
	}
}

func (s *ExcelService) ExportToExcel() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Funds
	fundsSheet := "Фонды"
	f.SetSheetName("Sheet1", fundsSheet)
	
	fundHeaders := []string{"ID", "Название", "ISIN", "Тикер", "УК", "Сегмент", "Квал", "ММ", "Дата создания"}
	for i, header := range fundHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(fundsSheet, cell, header)
	}

	funds, err := s.fundRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get funds: %w", err)
	}

	for row, fund := range funds {
		f.SetCellValue(fundsSheet, cellName(1, row+2), fund.ID)
		f.SetCellValue(fundsSheet, cellName(2, row+2), fund.Name)
		f.SetCellValue(fundsSheet, cellName(3, row+2), fund.ISIN)
		f.SetCellValue(fundsSheet, cellName(4, row+2), fund.Ticker)
		f.SetCellValue(fundsSheet, cellName(5, row+2), fund.ManagementCompany)
		f.SetCellValue(fundsSheet, cellName(6, row+2), fund.RealEstateSegment)
		f.SetCellValue(fundsSheet, cellName(7, row+2), boolToString(fund.QualifiedRequired))
		f.SetCellValue(fundsSheet, cellName(8, row+2), boolToString(fund.HasMarketMaker))
		f.SetCellValue(fundsSheet, cellName(9, row+2), fund.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Sheet 2: Financials
	financialsSheet := "Финансы"
	f.NewSheet(financialsSheet)
	
	financialHeaders := []string{
		"Fund ID", "Название фонда", "Дата среза", "Цена пая", "NAV", "Дисконт %",
		"Cap Rate %", "P/NAV", "P/AFFO", "NOI Yield %", "Выплата в год",
		"Доходность выплат %", "Полная доходность %", "Долг/СЧА", "Комиссия УК %",
		"Объём торгов", "Объектов", "IRR прогноз %",
	}
	for i, header := range financialHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(financialsSheet, cell, header)
	}

	row := 2
	for _, fund := range funds {
		financials, err := s.financialsRepo.GetByFundID(fund.ID)
		if err != nil || len(financials) == 0 {
			continue
		}

		latest := financials[0]
		f.SetCellValue(financialsSheet, cellName(1, row), latest.FundID)
		f.SetCellValue(financialsSheet, cellName(2, row), fund.Name)
		f.SetCellValue(financialsSheet, cellName(3, row), latest.SnapshotDate.Format("2006-01-02"))
		f.SetCellValue(financialsSheet, cellName(4, row), latest.UnitPriceRub)
		f.SetCellValue(financialsSheet, cellName(5, row), latest.NavPerUnitRub)
		f.SetCellValue(financialsSheet, cellName(6, row), latest.DiscountToNavPct)
		f.SetCellValue(financialsSheet, cellName(7, row), latest.CapRatePct)
		f.SetCellValue(financialsSheet, cellName(8, row), latest.PNav)
		f.SetCellValue(financialsSheet, cellName(9, row), latest.PAFFO)
		f.SetCellValue(financialsSheet, cellName(10, row), latest.NoiYieldPct)
		f.SetCellValue(financialsSheet, cellName(11, row), latest.AnnualPayoutRub)
		f.SetCellValue(financialsSheet, cellName(12, row), latest.PayoutYieldPct)
		f.SetCellValue(financialsSheet, cellName(13, row), latest.TotalReturnPct)
		f.SetCellValue(financialsSheet, cellName(14, row), latest.DebtToNavRatio)
		f.SetCellValue(financialsSheet, cellName(15, row), latest.ManagementFeePct)
		f.SetCellValue(financialsSheet, cellName(16, row), latest.TradingVolumeMlnRub)
		f.SetCellValue(financialsSheet, cellName(17, row), latest.NumberOfProperties)
		f.SetCellValue(financialsSheet, cellName(18, row), latest.IRRForecastPct)
		row++
	}

	// Sheet 3: Analysis
	if s.analysisRepo != nil {
		analysisSheet := "Анализ"
		f.NewSheet(analysisSheet)
		
		analysisHeaders := []string{"Fund ID", "Название фонда", "Модель", "Дата анализа", "Резюме", "Оценка рисков", "Плюсы/Минусы"}
		for i, header := range analysisHeaders {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(analysisSheet, cell, header)
		}

		row := 2
		for _, fund := range funds {
			analysis, err := s.analysisRepo.GetLatestByFundID(fund.ID)
			if err != nil || analysis == nil {
				continue
			}

			f.SetCellValue(analysisSheet, cellName(1, row), analysis.FundID)
			f.SetCellValue(analysisSheet, cellName(2, row), fund.Name)
			f.SetCellValue(analysisSheet, cellName(3, row), analysis.ModelUsed)
			f.SetCellValue(analysisSheet, cellName(4, row), analysis.CreatedAt.Format("2006-01-02 15:04:05"))
			f.SetCellValue(analysisSheet, cellName(5, row), analysis.AnalysisSummary)
			f.SetCellValue(analysisSheet, cellName(6, row), analysis.RiskAssessment)
			f.SetCellValue(analysisSheet, cellName(7, row), analysis.ProsCons)
			row++
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *ExcelService) ImportFromExcel(data []byte) (int, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("failed to open excel: %w", err)
	}
	defer f.Close()

	imported := 0

	// Import funds from "Фонды" sheet
	fundsSheet := "Фонды"
	if idx, _ := f.GetSheetIndex(fundsSheet); idx != -1 {
		rows, err := f.GetRows(fundsSheet)
		if err != nil {
			return 0, fmt.Errorf("failed to read funds sheet: %w", err)
		}

		for i, row := range rows {
			if i == 0 {
				continue // Skip header
			}
			if len(row) < 6 {
				continue
			}

			fund := &models.Fund{
				Name:              row[1],
				ISIN:              row[2],
				Ticker:            row[3],
				ManagementCompany: row[4],
				RealEstateSegment: row[5],
			}
			if len(row) > 6 {
				fund.QualifiedRequired = stringToBool(row[6])
			}
			if len(row) > 7 {
				fund.HasMarketMaker = stringToBool(row[7])
			}

			// Check if fund already exists by ISIN
			existing, _ := s.fundRepo.GetByISIN(fund.ISIN)
			if existing != nil {
				fund.ID = existing.ID
				if err := s.fundRepo.Update(fund); err != nil {
					return imported, fmt.Errorf("failed to update fund %s: %w", fund.ISIN, err)
				}
			} else {
				if err := s.fundRepo.Create(fund); err != nil {
					return imported, fmt.Errorf("failed to create fund %s: %w", fund.ISIN, err)
				}
			}
			imported++
		}
	}

	// Import financials from "Финансы" sheet
	financialsSheet := "Финансы"
	if idx, _ := f.GetSheetIndex(financialsSheet); idx != -1 {
		rows, err := f.GetRows(financialsSheet)
		if err != nil {
			return imported, fmt.Errorf("failed to read financials sheet: %w", err)
		}

		for i, row := range rows {
			if i == 0 {
				continue // Skip header
			}
			if len(row) < 4 {
				continue
			}

			fundID, err := strconv.ParseUint(row[0], 10, 32)
			if err != nil {
				continue
			}

			snapshotDate, err := time.Parse("2006-01-02", row[2])
			if err != nil {
				snapshotDate = time.Now()
			}

			financials := &models.FundFinancials{
				FundID:       uint(fundID),
				SnapshotDate: snapshotDate,
			}

			if len(row) > 3 {
				financials.UnitPriceRub = parseFloat(row[3])
			}
			if len(row) > 4 {
				financials.NavPerUnitRub = parseFloat(row[4])
			}
			if len(row) > 5 {
				financials.DiscountToNavPct = parseFloat(row[5])
			}
			if len(row) > 6 {
				financials.CapRatePct = parseFloat(row[6])
			}
			if len(row) > 7 {
				financials.PNav = parseFloat(row[7])
			}
			if len(row) > 8 {
				financials.PAFFO = parseFloat(row[8])
			}
			if len(row) > 9 {
				financials.NoiYieldPct = parseFloat(row[9])
			}
			if len(row) > 10 {
				financials.AnnualPayoutRub = parseFloat(row[10])
			}
			if len(row) > 11 {
				financials.PayoutYieldPct = parseFloat(row[11])
			}
			if len(row) > 12 {
				financials.TotalReturnPct = parseFloat(row[12])
			}
			if len(row) > 13 {
				financials.DebtToNavRatio = parseFloat(row[13])
			}
			if len(row) > 14 {
				financials.ManagementFeePct = parseFloat(row[14])
			}
			if len(row) > 15 {
				financials.TradingVolumeMlnRub = parseFloat(row[15])
			}
			if len(row) > 16 {
				financials.NumberOfProperties = parseInt(row[16])
			}
			if len(row) > 17 {
				financials.IRRForecastPct = parseFloat(row[17])
			}

			if err := s.financialsRepo.Create(financials); err != nil {
				return imported, fmt.Errorf("failed to create financials for fund %d: %w", fundID, err)
			}
			imported++
		}
	}

	return imported, nil
}

func (s *ExcelService) GetAllFundsData() ([]models.Fund, error) {
	return s.fundRepo.GetAll()
}

func (s *ExcelService) ImportFund(fund *models.Fund) error {
	return s.fundRepo.Create(fund)
}

func (s *ExcelService) ImportFinancials(financials *models.FundFinancials) error {
	return s.financialsRepo.Create(financials)
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func boolToString(b bool) string {
	if b {
		return "Да"
	}
	return "Нет"
}

func stringToBool(s string) bool {
	return s == "Да" || s == "да" || s == "Yes" || s == "yes" || s == "true" || s == "1"
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}
