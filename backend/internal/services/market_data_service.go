package services

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/parsers"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

type FetchResult struct {
	Status               string `json:"status"`
	FundID               uint   `json:"fund_id,omitempty"`
	RecordsCreated       int    `json:"records_created"`
	RecordsUpdated       int    `json:"records_updated"`
	MoexAvailable        bool   `json:"moex_available"`
	InvestfundsAvailable bool   `json:"investfunds_available"`
	Error                string `json:"error,omitempty"`
}

type MarketDataService struct {
	moexParser        *parsers.MoexParser
	investfundsParser *parsers.InvestfundsParser
	financialsRepo    *repositories.FinancialsRepository
	fundRepo          *repositories.FundRepository
}

func NewMarketDataService(
	moexParser *parsers.MoexParser,
	investfundsParser *parsers.InvestfundsParser,
	financialsRepo *repositories.FinancialsRepository,
	fundRepo *repositories.FundRepository,
) *MarketDataService {
	return &MarketDataService{
		moexParser:        moexParser,
		investfundsParser: investfundsParser,
		financialsRepo:    financialsRepo,
		fundRepo:          fundRepo,
	}
}

func (s *MarketDataService) FetchMarketDataForFund(ctx context.Context, fundID uint) (*FetchResult, error) {
	fund, err := s.fundRepo.GetByID(fundID)
	if err != nil {
		return nil, fmt.Errorf("fund not found: %w", err)
	}

	result := &FetchResult{
		Status: "success",
		FundID: fundID,
	}

	var moexHistory []parsers.MoexMarketData
	var investfundsData *parsers.InvestfundsData

	// Try to get MOEX data
	// If ticker is empty, search by ISIN
	if fund.ISIN != "" {
		log.Printf("Searching MOEX by ISIN: %s", fund.ISIN)
		security, err := s.moexParser.SearchSecurity(fund.ISIN)
		if err != nil {
			log.Printf("MOEX search error for ISIN %s: %v", fund.ISIN, err)
		} else {
			log.Printf("MOEX: found security %s with boards: %v", security.SecID, security.Boards)

			// Получаем данные со всех board'ов
			for _, board := range security.Boards {
				log.Printf("Fetching MOEX data for board: %s", board)
				history, err := s.moexParser.GetPriceHistoryWithBoard(security.SecID, board)
				if err != nil {
					log.Printf("MOEX error for board %s: %v", board, err)
					continue
				}
				moexHistory = append(moexHistory, history...)
				log.Printf("MOEX: fetched %d records from board %s", len(history), board)
			}

			if len(moexHistory) > 0 {
				result.MoexAvailable = true
				log.Printf("MOEX: total %d records from all boards", len(moexHistory))
			}
		}
	} else if fund.Ticker != "" {
		// Fallback: используем ticker с board TQBR
		log.Printf("Fetching MOEX data for ticker: %s", fund.Ticker)
		history, err := s.moexParser.GetPriceHistoryWithBoard(fund.Ticker, "TQBR")
		if err != nil {
			log.Printf("MOEX error for ticker %s: %v", fund.Ticker, err)
		} else {
			moexHistory = history
			result.MoexAvailable = true
			log.Printf("MOEX: fetched %d records for ticker %s", len(history), fund.Ticker)
		}
	} else {
		log.Printf("No ticker or ISIN for fund %s, skipping MOEX", fund.Name)
	}

	// Try to get investfunds data
	var fundURL string

	if fund.InvestfundsURL != "" {
		// Use provided URL directly
		fundURL = fund.InvestfundsURL
		log.Printf("Using provided investfunds URL: %s", fundURL)
	} else {
		// Try to search for fund
		searchQuery := fund.ISIN
		if searchQuery == "" || len(searchQuery) > 12 {
			searchQuery = fund.Name
		}

		log.Printf("Searching investfunds for: %s", searchQuery)
		fundURL, err = s.investfundsParser.SearchFund(searchQuery)
		if err != nil {
			log.Printf("investfunds search error for %s: %v", searchQuery, err)
			result.InvestfundsAvailable = false
		}
	}

	if fundURL != "" {
		data, err := s.investfundsParser.GetFundData(fundURL)
		if err != nil {
			log.Printf("investfunds data error for %s: %v", fundURL, err)
			result.InvestfundsAvailable = false
		} else {
			investfundsData = data
			result.InvestfundsAvailable = true
			log.Printf("investfunds: fetched data for %s (NAV: %.2f, NAV history: %d, Payouts: %d)",
				fund.Name, data.NAV, len(data.NAVHistory), len(data.PayoutHistory))
		}
	}

	if len(moexHistory) == 0 && investfundsData == nil {
		return nil, fmt.Errorf("no data available from any source")
	}

	// Process MOEX price history (создаёт записи с ценой)
	for _, moexData := range moexHistory {
		financials := &models.FundFinancials{
			FundID:       fundID,
			SnapshotDate: moexData.Date,
			UnitPriceRub: moexData.Close,
		}

		existing, _ := s.financialsRepo.GetByFundIDAndDate(fundID, moexData.Date)
		if existing != nil {
			existing.UnitPriceRub = financials.UnitPriceRub
			if err := s.financialsRepo.Update(existing); err != nil {
				log.Printf("Failed to update price for date %s: %v", moexData.Date, err)
			} else {
				result.RecordsUpdated++
			}
		} else {
			if err := s.financialsRepo.Create(financials); err != nil {
				log.Printf("Failed to create price for date %s: %v", moexData.Date, err)
			} else {
				result.RecordsCreated++
			}
		}
	}

	// Process investfunds NAV history (ВСЕГДА, объединяя с MOEX данными)
	if investfundsData != nil && len(investfundsData.NAVHistory) > 0 {
		for _, navData := range investfundsData.NAVHistory {
			existing, _ := s.financialsRepo.GetByFundIDAndDate(fundID, navData.Date)
			if existing != nil {
				// Обновляем существующую запись (добавляем NAV и СЧА)
				existing.NavPerUnitRub = navData.NAV
				existing.NavTotalMlnRub = navData.SCA / 1_000_000
				
				// НЕ устанавливаем UnitPrice из investfunds — это не рыночная цена
				
				// Пересчитываем дисконт, если есть цена из MOEX
				if existing.UnitPriceRub > 0 && navData.NAV > 0 {
					existing.DiscountToNavPct = ((existing.UnitPriceRub - navData.NAV) / navData.NAV) * 100
				}
				if err := s.financialsRepo.Update(existing); err != nil {
					log.Printf("Failed to update NAV for date %s: %v", navData.Date, err)
				} else {
					result.RecordsUpdated++
				}
			} else {
				// Создаём новую запись с NAV и СЧА (без цены)
				financials := &models.FundFinancials{
					FundID:         fundID,
					SnapshotDate:   navData.Date,
					UnitPriceRub:   0,               // НЕ устанавливать
					NavPerUnitRub:  navData.NAV,     // РСП
					NavTotalMlnRub: navData.SCA / 1_000_000,  // СЧА
				}
				if err := s.financialsRepo.Create(financials); err != nil {
					log.Printf("Failed to create NAV for date %s: %v", navData.Date, err)
				} else {
					result.RecordsCreated++
				}
			}
		}
	}

	// Process payout history
	if investfundsData != nil && len(investfundsData.PayoutHistory) > 0 {
		for _, payout := range investfundsData.PayoutHistory {
			existing, _ := s.financialsRepo.GetByFundIDAndDate(fundID, payout.Date)
			if existing != nil {
				existing.AnnualPayoutRub = payout.Amount
				existing.PayoutYieldPct = payout.YieldPercent
				if err := s.financialsRepo.Update(existing); err != nil {
					log.Printf("Failed to update payout for date %s: %v", payout.Date, err)
				}
			} else {
				financials := &models.FundFinancials{
					FundID:           fundID,
					SnapshotDate:     payout.Date,
					AnnualPayoutRub:  payout.Amount,
					PayoutYieldPct:   payout.YieldPercent,
				}
				if err := s.financialsRepo.Create(financials); err != nil {
					log.Printf("Failed to create payout for date %s: %v", payout.Date, err)
				} else {
					result.RecordsCreated++
				}
			}
		}
	}

	// Интерполяция пропусков в NAV
	if err := s.interpolateNAV(fundID); err != nil {
		log.Printf("Failed to interpolate NAV for fund %d: %v", fundID, err)
	}

	return result, nil
}

func (s *MarketDataService) interpolateNAV(fundID uint) error {
	// Получить все записи для фонда
	financials, err := s.financialsRepo.GetByFundID(fundID)
	if err != nil {
		return err
	}

	if len(financials) < 2 {
		return nil
	}

	// Сортировка по дате
	sort.Slice(financials, func(i, j int) bool {
		return financials[i].SnapshotDate.Before(financials[j].SnapshotDate)
	})

	// Интерполяция пропусков
	for i := range financials {
		if financials[i].NavPerUnitRub == 0 {
			// Найти предыдущее значение
			prevNAV := 0.0
			prevDate := time.Time{}
			for j := i - 1; j >= 0; j-- {
				if financials[j].NavPerUnitRub > 0 {
					prevNAV = financials[j].NavPerUnitRub
					prevDate = financials[j].SnapshotDate
					break
				}
			}

			// Найти следующее значение
			nextNAV := 0.0
			nextDate := time.Time{}
			for j := i + 1; j < len(financials); j++ {
				if financials[j].NavPerUnitRub > 0 {
					nextNAV = financials[j].NavPerUnitRub
					nextDate = financials[j].SnapshotDate
					break
				}
			}

			// Интерполяция или extrapolation
			if prevNAV > 0 && nextNAV > 0 {
				totalDays := nextDate.Sub(prevDate).Hours() / 24
				currentDays := financials[i].SnapshotDate.Sub(prevDate).Hours() / 24
				ratio := currentDays / totalDays
				financials[i].NavPerUnitRub = prevNAV + (nextNAV-prevNAV)*ratio

				// Пересчитать дисконт
				if financials[i].UnitPriceRub > 0 {
					financials[i].DiscountToNavPct = ((financials[i].UnitPriceRub - financials[i].NavPerUnitRub) / financials[i].NavPerUnitRub) * 100
				}

				s.financialsRepo.Update(&financials[i])
			} else if prevNAV > 0 {
				// Extrapolation: использовать последнее известное значение
				financials[i].NavPerUnitRub = prevNAV

				if financials[i].UnitPriceRub > 0 {
					financials[i].DiscountToNavPct = ((financials[i].UnitPriceRub - prevNAV) / prevNAV) * 100
				}

				s.financialsRepo.Update(&financials[i])
			}
		}
	}

	return nil
}

func (s *MarketDataService) FetchMarketDataForAllFunds(ctx context.Context) (*FetchResult, error) {
	funds, err := s.fundRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get funds: %w", err)
	}

	totalResult := &FetchResult{
		Status: "success",
	}

	for _, fund := range funds {
		log.Printf("Processing fund: %s (ID: %d)", fund.Name, fund.ID)
		result, err := s.FetchMarketDataForFund(ctx, fund.ID)
		if err != nil {
			log.Printf("Error processing fund %s: %v", fund.Name, err)
			continue
		}

		totalResult.RecordsCreated += result.RecordsCreated
		totalResult.RecordsUpdated += result.RecordsUpdated

		if result.MoexAvailable {
			totalResult.MoexAvailable = true
		}
		if result.InvestfundsAvailable {
			totalResult.InvestfundsAvailable = true
		}

		time.Sleep(500 * time.Millisecond)
	}

	return totalResult, nil
}
