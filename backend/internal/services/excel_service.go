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

	sheetName := "Фонды"
	f.SetSheetName("Sheet1", sheetName)

	rubFmt := `#,##0.00" "[$₽-419]`
	pctFmt := "[$-419]0.00%"
	numFmt := "[$-419]0.00"
	rubStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &rubFmt})
	pctStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &pctFmt})
	numStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &numFmt})

	headers := []string{
		"Название", "ISIN", "Тикер", "Квал", "УК", "Сегмент",
		"Цена пая", "РСП", "Дисконт к РСП", "Cap Rate", "СЧА, млн ₽",
		"P/NAV", "P/AFFO", "Доходность выплат", "Комиссия УК",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	funds, err := s.fundRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get funds: %w", err)
	}

	for row, fund := range funds {
		r := row + 2
		f.SetCellValue(sheetName, cellName(1, r), fund.Name)
		f.SetCellValue(sheetName, cellName(2, r), fund.ISIN)
		f.SetCellValue(sheetName, cellName(3, r), fund.Ticker)
		f.SetCellValue(sheetName, cellName(4, r), boolToString(fund.QualifiedRequired))
		f.SetCellValue(sheetName, cellName(5, r), fund.ManagementCompany)
		f.SetCellValue(sheetName, cellName(6, r), fund.RealEstateSegment)

		financials, err := s.financialsRepo.GetByFundID(fund.ID)
		if err != nil || len(financials) == 0 {
			continue
		}

		latest := financials[0]
		f.SetCellValue(sheetName, cellName(7, r), latest.UnitPriceRub)
		f.SetCellValue(sheetName, cellName(8, r), latest.NavPerUnitRub)
		f.SetCellValue(sheetName, cellName(9, r), latest.DiscountToNavPct/100)
		f.SetCellValue(sheetName, cellName(10, r), latest.CapRatePct/100)
		f.SetCellValue(sheetName, cellName(11, r), latest.NavTotalMlnRub)
		f.SetCellValue(sheetName, cellName(12, r), latest.PNav)
		f.SetCellValue(sheetName, cellName(13, r), latest.PAFFO)
		f.SetCellValue(sheetName, cellName(14, r), latest.PayoutYieldPct/100)
		f.SetCellValue(sheetName, cellName(15, r), latest.ManagementFeePct/100)
	}

	lastRow := len(funds) + 1
	if lastRow >= 2 {
		f.SetCellStyle(sheetName, "G2", cellName(7, lastRow), rubStyle)
		f.SetCellStyle(sheetName, "H2", cellName(8, lastRow), rubStyle)
		f.SetCellStyle(sheetName, "I2", cellName(9, lastRow), pctStyle)
		f.SetCellStyle(sheetName, "J2", cellName(10, lastRow), pctStyle)
		f.SetCellStyle(sheetName, "K2", cellName(11, lastRow), rubStyle)
		f.SetCellStyle(sheetName, "L2", cellName(12, lastRow), numStyle)
		f.SetCellStyle(sheetName, "M2", cellName(13, lastRow), numStyle)
		f.SetCellStyle(sheetName, "N2", cellName(14, lastRow), pctStyle)
		f.SetCellStyle(sheetName, "O2", cellName(15, lastRow), pctStyle)
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
