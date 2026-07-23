package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zpif-analyzer/backend/internal/llm"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

type Discoverer interface {
	Discover(ctx context.Context, fund *models.Fund) (*llm.DiscoveryStatus, error)
	GetStatus(fundID uint) *llm.DiscoveryStatus
}

type Analyzer interface {
	AnalyzeLatestDocuments(ctx context.Context, fund *models.Fund) (*models.LLMAnalysis, error)
}

type FundService struct {
	fundRepo       *repositories.FundRepository
	financialsRepo *repositories.FinancialsRepository
	documentRepo   *repositories.DocumentRepository
	analysisRepo   *repositories.AnalysisRepository
	llmSettingsRepo *repositories.LLMSettingsRepository
	discoverer     Discoverer
	analyzer       Analyzer
}

func NewFundService(
	fundRepo *repositories.FundRepository,
	financialsRepo *repositories.FinancialsRepository,
	documentRepo *repositories.DocumentRepository,
	analysisRepo *repositories.AnalysisRepository,
) *FundService {
	return &FundService{
		fundRepo:       fundRepo,
		financialsRepo: financialsRepo,
		documentRepo:   documentRepo,
		analysisRepo:   analysisRepo,
	}
}

func (s *FundService) SetDiscoverer(d Discoverer) {
	s.discoverer = d
}

func (s *FundService) SetAnalyzer(a Analyzer) {
	s.analyzer = a
}

func (s *FundService) SetLLMSettingsRepo(repo *repositories.LLMSettingsRepository) {
	s.llmSettingsRepo = repo
}

func (s *FundService) GetAllFunds() ([]models.Fund, error) {
	return s.fundRepo.GetAll()
}

func (s *FundService) GetFundByID(id uint) (*models.Fund, error) {
	fund, err := s.fundRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return fund, nil
}

func (s *FundService) CreateFund(fund *models.Fund) error {
	// Validate ISIN format (basic validation)
	if fund.ISIN == "" {
		return errors.New("ISIN is required")
	}
	
	// Check if ISIN already exists
	existing, _ := s.fundRepo.GetByISIN(fund.ISIN)
	if existing != nil {
		return fmt.Errorf("fund with ISIN %s already exists", fund.ISIN)
	}
	
	return s.fundRepo.Create(fund)
}

func (s *FundService) UpdateFund(id uint, fund *models.Fund) error {
	existing, err := s.fundRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	// Update fields
	existing.Name = fund.Name
	existing.ISIN = fund.ISIN
	existing.Ticker = fund.Ticker
	existing.ManagementCompany = fund.ManagementCompany
	existing.RealEstateSegment = fund.RealEstateSegment
	existing.QualifiedRequired = fund.QualifiedRequired
	existing.HasMarketMaker = fund.HasMarketMaker
	existing.FundEndDate = fund.FundEndDate
	existing.InvestfundsURL = fund.InvestfundsURL
	existing.VsezpifURL = fund.VsezpifURL
	
	return s.fundRepo.Update(existing)
}

func (s *FundService) DeleteFund(id uint) error {
	return s.fundRepo.Delete(id)
}

func (s *FundService) GetFinancialsByFundID(fundID uint) ([]models.FundFinancials, error) {
	return s.financialsRepo.GetByFundID(fundID)
}

func (s *FundService) GetLatestFinancials(fundID uint) (*models.FundFinancials, error) {
	return s.financialsRepo.GetLatestByFundID(fundID)
}

func (s *FundService) AddFinancials(fundID uint, financials *models.FundFinancials) error {
	// Verify fund exists
	_, err := s.fundRepo.GetByID(fundID)
	if err != nil {
		return err
	}
	
	financials.FundID = fundID
	return s.financialsRepo.Create(financials)
}

func (s *FundService) GetDocumentsByFundID(fundID uint) ([]models.FundDocument, error) {
	return s.documentRepo.GetByFundID(fundID)
}

func (s *FundService) AddDocument(document *models.FundDocument) error {
	// Verify fund exists
	_, err := s.fundRepo.GetByID(document.FundID)
	if err != nil {
		return err
	}
	
	// Check if document with same hash already exists
	existing, _ := s.documentRepo.GetByHash(document.ContentHash)
	if existing != nil {
		return errors.New("document with same content already exists")
	}
	
	return s.documentRepo.Create(document)
}

func (s *FundService) DeleteDocument(documentID uint) error {
	return s.documentRepo.Delete(documentID)
}

func (s *FundService) GetDocumentByID(documentID uint) (*models.FundDocument, error) {
	return s.documentRepo.GetByID(documentID)
}

func (s *FundService) GetLatestAnalysis(fundID uint) (*models.LLMAnalysis, error) {
	return s.analysisRepo.GetLatestByFundID(fundID)
}

func (s *FundService) AddAnalysis(analysis *models.LLMAnalysis) error {
	return s.analysisRepo.Create(analysis)
}

func (s *FundService) DiscoverDocumentsForFund(fundID uint) error {
	if s.discoverer == nil {
		return errors.New("document discovery not configured")
	}
	fund, err := s.fundRepo.GetByID(fundID)
	if err != nil {
		return err
	}
	_, err = s.discoverer.Discover(context.Background(), fund)
	return err
}

func (s *FundService) DiscoverDocumentsForAllFunds() error {
	if s.discoverer == nil {
		return errors.New("document discovery not configured")
	}
	funds, err := s.fundRepo.GetAll()
	if err != nil {
		return err
	}
	var lastErr error
	for i := range funds {
		if _, err := s.discoverer.Discover(context.Background(), &funds[i]); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (s *FundService) AnalyzeFund(ctx context.Context, fundID uint) (*models.LLMAnalysis, error) {
	if s.analyzer == nil {
		return nil, errors.New("analyzer not configured")
	}
	fund, err := s.fundRepo.GetByID(fundID)
	if err != nil {
		return nil, err
	}
	return s.analyzer.AnalyzeLatestDocuments(ctx, fund)
}

func (s *FundService) GetDiscoveryStatus(fundID uint) map[string]interface{} {
	if s.discoverer == nil {
		return map[string]interface{}{"status": "idle", "fund_id": fundID}
	}
	status := s.discoverer.GetStatus(fundID)
	return map[string]interface{}{
		"status":  status.Status,
		"fund_id": status.FundID,
	}
}

func (s *FundService) EnrichAndCreateFund(ctx context.Context, userInput string) (*models.Fund, error) {
	if s.llmSettingsRepo == nil {
		return nil, errors.New("LLM settings not configured")
	}

	settings, err := s.llmSettingsRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load LLM settings: %w", err)
	}

	if settings.APIKeyEncrypted == "" {
		return nil, errors.New("LLM API key not configured")
	}

	client := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.ModelName)

	response, err := client.ChatSimple(ctx, llm.EnrichFundPrompt, userInput)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	jsonStr := extractJSON(response)

	var enrichedData struct {
		Name              string  `json:"name"`
		ISIN              string  `json:"isin"`
		Ticker            string  `json:"ticker"`
		ManagementCompany string  `json:"management_company"`
		RealEstateSegment string  `json:"real_estate_segment"`
		QualifiedRequired bool    `json:"qualified_required"`
		HasMarketMaker    bool    `json:"has_market_maker"`
		FundEndDate       *string `json:"fund_end_date"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &enrichedData); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if enrichedData.ISIN == "" || enrichedData.ISIN == "UNKNOWN" {
		enrichedData.ISIN = fmt.Sprintf("PENDING-%d", time.Now().Unix())
	}

	fund := &models.Fund{
		Name:              enrichedData.Name,
		ISIN:              enrichedData.ISIN,
		Ticker:            enrichedData.Ticker,
		ManagementCompany: enrichedData.ManagementCompany,
		RealEstateSegment: enrichedData.RealEstateSegment,
		QualifiedRequired: enrichedData.QualifiedRequired,
		HasMarketMaker:    enrichedData.HasMarketMaker,
	}

	if enrichedData.FundEndDate != nil && *enrichedData.FundEndDate != "" {
		if t, err := time.Parse("2006-01-02", *enrichedData.FundEndDate); err == nil {
			fund.FundEndDate = &t
		}
	}

	if err := s.CreateFund(fund); err != nil {
		return nil, err
	}

	return fund, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}
