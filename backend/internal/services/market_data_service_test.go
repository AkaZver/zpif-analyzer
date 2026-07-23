package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/parsers"
)

type MockMoexParser struct {
	mock.Mock
}

func (m *MockMoexParser) SearchSecurity(isin string) (*parsers.MoexSecurity, error) {
	args := m.Called(isin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*parsers.MoexSecurity), args.Error(1)
}

func (m *MockMoexParser) GetPriceHistoryWithBoard(secID, board string) ([]parsers.MoexMarketData, error) {
	args := m.Called(secID, board)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]parsers.MoexMarketData), args.Error(1)
}

type MockInvestfundsParser struct {
	mock.Mock
}

func (m *MockInvestfundsParser) SearchFund(query string) (string, error) {
	args := m.Called(query)
	return args.String(0), args.Error(1)
}

func (m *MockInvestfundsParser) GetFundData(fundURL string) (*parsers.InvestfundsData, error) {
	args := m.Called(fundURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*parsers.InvestfundsData), args.Error(1)
}

type MockVsezpifParser struct {
	mock.Mock
}

func (m *MockVsezpifParser) GetFundDataByISIN(isin string) (*parsers.VsezpifData, error) {
	args := m.Called(isin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*parsers.VsezpifData), args.Error(1)
}

func (m *MockVsezpifParser) GetFundDataByURL(fundURL string) (*parsers.VsezpifData, error) {
	args := m.Called(fundURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*parsers.VsezpifData), args.Error(1)
}

type MockFinancialsRepo struct {
	mock.Mock
}

func (m *MockFinancialsRepo) GetByFundID(fundID uint) ([]models.FundFinancials, error) {
	args := m.Called(fundID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.FundFinancials), args.Error(1)
}

func (m *MockFinancialsRepo) GetByFundIDAndDate(fundID uint, date time.Time) (*models.FundFinancials, error) {
	args := m.Called(fundID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FundFinancials), args.Error(1)
}

func (m *MockFinancialsRepo) GetLatestByFundID(fundID uint) (*models.FundFinancials, error) {
	args := m.Called(fundID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FundFinancials), args.Error(1)
}

func (m *MockFinancialsRepo) Create(financial *models.FundFinancials) error {
	args := m.Called(financial)
	return args.Error(0)
}

func (m *MockFinancialsRepo) Update(financial *models.FundFinancials) error {
	args := m.Called(financial)
	return args.Error(0)
}

type MockFundRepo struct {
	mock.Mock
}

func (m *MockFundRepo) GetAll() ([]models.Fund, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Fund), args.Error(1)
}

func (m *MockFundRepo) GetByID(id uint) (*models.Fund, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Fund), args.Error(1)
}

func (m *MockFundRepo) Update(fund *models.Fund) error {
	args := m.Called(fund)
	return args.Error(0)
}

func TestMarketDataService_FetchMarketDataForFund_FundNotFound(t *testing.T) {
	fundRepo := new(MockFundRepo)
	financialsRepo := new(MockFinancialsRepo)
	moexParser := new(MockMoexParser)
	investfundsParser := new(MockInvestfundsParser)
	vsezpifParser := new(MockVsezpifParser)

	fundRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found"))

	service := &MarketDataService{
		moexParser:        moexParser,
		investfundsParser: investfundsParser,
		vsezpifParser:     vsezpifParser,
		financialsRepo:    financialsRepo,
		fundRepo:          fundRepo,
	}

	result, err := service.FetchMarketDataForFund(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "fund not found")
}

func TestMarketDataService_FetchMarketDataForFund_NoDataAvailable(t *testing.T) {
	fundRepo := new(MockFundRepo)
	financialsRepo := new(MockFinancialsRepo)
	moexParser := new(MockMoexParser)
	investfundsParser := new(MockInvestfundsParser)
	vsezpifParser := new(MockVsezpifParser)

	fund := &models.Fund{ID: 1, Name: "Test Fund", ISIN: "RU000TEST"}
	fundRepo.On("GetByID", uint(1)).Return(fund, nil)
	moexParser.On("SearchSecurity", "RU000TEST").Return(nil, errors.New("not found"))
	investfundsParser.On("SearchFund", "RU000TEST").Return("", errors.New("not found"))
	vsezpifParser.On("GetFundDataByISIN", "RU000TEST").Return(nil, errors.New("not found"))

	service := &MarketDataService{
		moexParser:        moexParser,
		investfundsParser: investfundsParser,
		vsezpifParser:     vsezpifParser,
		financialsRepo:    financialsRepo,
		fundRepo:          fundRepo,
	}

	result, err := service.FetchMarketDataForFund(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no data available")
}

func TestMarketDataService_FetchMarketDataForFund_WithMoexData(t *testing.T) {
	fundRepo := new(MockFundRepo)
	financialsRepo := new(MockFinancialsRepo)
	moexParser := new(MockMoexParser)
	investfundsParser := new(MockInvestfundsParser)
	vsezpifParser := new(MockVsezpifParser)

	fund := &models.Fund{ID: 1, Name: "Test Fund", ISIN: "RU000TEST"}
	fundRepo.On("GetByID", uint(1)).Return(fund, nil)

	security := &parsers.MoexSecurity{SecID: "TEST", Boards: []string{"TQBR"}, ISIN: "RU000TEST"}
	moexParser.On("SearchSecurity", "RU000TEST").Return(security, nil)

	moexHistory := []parsers.MoexMarketData{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Close: 1000.0},
	}
	moexParser.On("GetPriceHistoryWithBoard", "TEST", "TQBR").Return(moexHistory, nil)

	investfundsParser.On("SearchFund", "RU000TEST").Return("", errors.New("not found"))
	vsezpifParser.On("GetFundDataByISIN", "RU000TEST").Return(nil, errors.New("not found"))

	financialsRepo.On("GetByFundIDAndDate", uint(1), mock.Anything).Return(nil, errors.New("not found"))
	financialsRepo.On("Create", mock.AnythingOfType("*models.FundFinancials")).Return(nil)
	financialsRepo.On("GetByFundID", uint(1)).Return([]models.FundFinancials{}, nil)

	service := &MarketDataService{
		moexParser:        moexParser,
		investfundsParser: investfundsParser,
		vsezpifParser:     vsezpifParser,
		financialsRepo:    financialsRepo,
		fundRepo:          fundRepo,
	}

	result, err := service.FetchMarketDataForFund(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.True(t, result.MoexAvailable)
	assert.Equal(t, 1, result.RecordsCreated)
}

func TestMarketDataService_CalculateAnnualPayout(t *testing.T) {
	service := &MarketDataService{}

	payouts := []parsers.Payout{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Amount: 80.0},
		{Date: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), Amount: 75.0},
		{Date: time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC), Amount: 70.0},
		{Date: time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC), Amount: 100.0},
	}

	total := service.calculateAnnualPayout(payouts)

	assert.Equal(t, 225.0, total)
}

func TestMarketDataService_CalculateAnnualPayout_Empty(t *testing.T) {
	service := &MarketDataService{}

	total := service.calculateAnnualPayout([]parsers.Payout{})

	assert.Equal(t, 0.0, total)
}

func TestMarketDataService_InterpolateNAV(t *testing.T) {
	financialsRepo := new(MockFinancialsRepo)

	financials := []models.FundFinancials{
		{ID: 1, SnapshotDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), NavPerUnitRub: 1000.0, UnitPriceRub: 950.0},
		{ID: 2, SnapshotDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), NavPerUnitRub: 0, UnitPriceRub: 960.0},
		{ID: 3, SnapshotDate: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), NavPerUnitRub: 1100.0, UnitPriceRub: 1050.0},
	}

	financialsRepo.On("GetByFundID", uint(1)).Return(financials, nil)
	financialsRepo.On("Update", mock.AnythingOfType("*models.FundFinancials")).Return(nil)

	service := &MarketDataService{
		financialsRepo: financialsRepo,
	}

	err := service.interpolateNAV(1)

	assert.NoError(t, err)
}

func TestMarketDataService_InterpolateNAV_LessThanTwoRecords(t *testing.T) {
	financialsRepo := new(MockFinancialsRepo)

	financials := []models.FundFinancials{
		{ID: 1, SnapshotDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), NavPerUnitRub: 1000.0},
	}

	financialsRepo.On("GetByFundID", uint(1)).Return(financials, nil)

	service := &MarketDataService{
		financialsRepo: financialsRepo,
	}

	err := service.interpolateNAV(1)

	assert.NoError(t, err)
}

func TestMarketDataService_CalculateDerivedMetrics(t *testing.T) {
	financialsRepo := new(MockFinancialsRepo)

	financials := []models.FundFinancials{
		{
			ID:                1,
			UnitPriceRub:      1000.0,
			NavPerUnitRub:     1050.0,
			AnnualPayoutRub:   80.0,
			ManagementFeePct:  1.5,
		},
	}

	financialsRepo.On("GetByFundID", uint(1)).Return(financials, nil)
	financialsRepo.On("Update", mock.AnythingOfType("*models.FundFinancials")).Return(nil)

	service := &MarketDataService{
		financialsRepo: financialsRepo,
	}

	err := service.calculateDerivedMetrics(1)

	assert.NoError(t, err)
}

func TestMarketDataService_FetchMarketDataForAllFunds(t *testing.T) {
	fundRepo := new(MockFundRepo)
	financialsRepo := new(MockFinancialsRepo)
	moexParser := new(MockMoexParser)
	investfundsParser := new(MockInvestfundsParser)
	vsezpifParser := new(MockVsezpifParser)

	funds := []models.Fund{
		{ID: 1, Name: "Fund 1", ISIN: "RU0001"},
	}
	fundRepo.On("GetAll").Return(funds, nil)
	fundRepo.On("GetByID", uint(1)).Return(&funds[0], nil)
	moexParser.On("SearchSecurity", "RU0001").Return(nil, errors.New("not found"))
	investfundsParser.On("SearchFund", "RU0001").Return("", errors.New("not found"))
	vsezpifParser.On("GetFundDataByISIN", "RU0001").Return(nil, errors.New("not found"))

	service := &MarketDataService{
		moexParser:        moexParser,
		investfundsParser: investfundsParser,
		vsezpifParser:     vsezpifParser,
		financialsRepo:    financialsRepo,
		fundRepo:          fundRepo,
	}

	result, err := service.FetchMarketDataForAllFunds(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
}

func TestMarketDataService_FetchMarketDataForAllFunds_GetAllError(t *testing.T) {
	fundRepo := new(MockFundRepo)
	financialsRepo := new(MockFinancialsRepo)
	moexParser := new(MockMoexParser)
	investfundsParser := new(MockInvestfundsParser)
	vsezpifParser := new(MockVsezpifParser)

	fundRepo.On("GetAll").Return(nil, errors.New("db error"))

	service := &MarketDataService{
		moexParser:        moexParser,
		investfundsParser: investfundsParser,
		vsezpifParser:     vsezpifParser,
		financialsRepo:    financialsRepo,
		fundRepo:          fundRepo,
	}

	result, err := service.FetchMarketDataForAllFunds(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
}
