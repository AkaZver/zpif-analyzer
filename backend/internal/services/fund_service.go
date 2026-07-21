package services

import (
	"context"
	"errors"
	"fmt"

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
	existing.Ticker = fund.Ticker
	existing.ManagementCompany = fund.ManagementCompany
	existing.RealEstateSegment = fund.RealEstateSegment
	existing.QualifiedRequired = fund.QualifiedRequired
	existing.HasMarketMaker = fund.HasMarketMaker
	existing.FundStartDate = fund.FundStartDate
	existing.FundEndDate = fund.FundEndDate
	
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
