package services

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
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
