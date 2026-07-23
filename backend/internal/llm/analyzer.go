package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

type Analyzer struct {
	settingsRepo    *repositories.LLMSettingsRepository
	documentRepo    *repositories.DocumentRepository
	analysisRepo    *repositories.AnalysisRepository
	financialsRepo   *repositories.FinancialsRepository
	fundRepo        *repositories.FundRepository
}

func NewAnalyzer(
	settingsRepo *repositories.LLMSettingsRepository,
	documentRepo *repositories.DocumentRepository,
	analysisRepo *repositories.AnalysisRepository,
	financialsRepo *repositories.FinancialsRepository,
	fundRepo *repositories.FundRepository,
) *Analyzer {
	return &Analyzer{
		settingsRepo:   settingsRepo,
		documentRepo:   documentRepo,
		analysisRepo:   analysisRepo,
		financialsRepo: financialsRepo,
		fundRepo:       fundRepo,
	}
}

type AnalysisResult struct {
	Summary        string   `json:"summary"`
	RiskAssessment string   `json:"risk_assessment"`
	Pros           []string `json:"pros"`
	Cons           []string `json:"cons"`
}

type MetricsExtraction struct {
	UnitPriceRub        *float64 `json:"unit_price_rub,omitempty"`
	NavPerUnitRub       *float64 `json:"nav_per_unit_rub,omitempty"`
	NavTotalMlnRub      *float64 `json:"nav_total_mln_rub,omitempty"`
	DiscountToNavPct    *float64 `json:"discount_to_nav_pct,omitempty"`
	CapRatePct          *float64 `json:"cap_rate_pct,omitempty"`
	PNav                *float64 `json:"p_nav,omitempty"`
	PAFFO               *float64 `json:"p_affo,omitempty"`
	NoiYieldPct         *float64 `json:"noi_yield_pct,omitempty"`
	AnnualPayoutRub     *float64 `json:"annual_payout_rub,omitempty"`
	PayoutYieldPct      *float64 `json:"payout_yield_pct,omitempty"`
	PayoutFrequency     string   `json:"payout_frequency,omitempty"`
	ManagementFeePct    *float64 `json:"management_fee_pct,omitempty"`
	TradingVolumeMlnRub *float64 `json:"trading_volume_mln_rub,omitempty"`
	NumberOfProperties  *int     `json:"number_of_properties,omitempty"`
}

const maxDocumentTextBytes = 16000

func (a *Analyzer) Analyze(ctx context.Context, fund *models.Fund, documentID uint) (*models.LLMAnalysis, error) {
	settings, err := a.settingsRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM settings: %w", err)
	}

	llmClient := NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.ModelName)

	document, err := a.documentRepo.GetByID(documentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	docText, err := a.readDocumentText(document)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}
	if len(docText) == 0 {
		return nil, fmt.Errorf("document is empty")
	}
	if len(docText) > maxDocumentTextBytes {
		docText = docText[:maxDocumentTextBytes] + "..."
	}

	metrics, err := a.extractMetrics(ctx, llmClient, docText)
	if err != nil {
		return nil, fmt.Errorf("metrics extraction failed: %w", err)
	}

	analysis, err := a.generateAnalysis(ctx, llmClient, docText, fund)
	if err != nil {
		return nil, fmt.Errorf("analysis generation failed: %w", err)
	}

	metricsJSON, _ := json.Marshal(metrics)
	prosConsJSON, _ := json.Marshal(struct {
		Pros []string `json:"pros"`
		Cons []string `json:"cons"`
	}{analysis.Pros, analysis.Cons})

	record := &models.LLMAnalysis{
		FundID:           fund.ID,
		DocumentID:       document.ID,
		ModelUsed:        llmClient.GetModel(),
		RawResponse:      docText,
		AnalysisSummary:  analysis.Summary,
		RiskAssessment:   analysis.RiskAssessment,
		ProsCons:         string(prosConsJSON),
		ExtractedMetrics: string(metricsJSON),
	}

	if err := a.analysisRepo.Create(record); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	if err := a.updateFinancialsFromMetrics(fund.ID, metrics); err != nil {
		return record, fmt.Errorf("analysis saved, but financials update failed: %v", err)
	}

	if err := a.documentRepo.UpdateStatus(document.ID, "analyzed"); err != nil {
		return record, fmt.Errorf("analysis saved, but status update failed: %v", err)
	}

	return record, nil
}

func (a *Analyzer) AnalyzeLatestDocuments(ctx context.Context, fund *models.Fund) (*models.LLMAnalysis, error) {
	docs, err := a.documentRepo.GetByFundID(fund.ID)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents to analyze")
	}

	var latest *models.FundDocument
	for i := range docs {
		if docs[i].Status != "analyzed" {
			latest = &docs[i]
			break
		}
	}
	if latest == nil {
		latest = &docs[0]
	}
	return a.Analyze(ctx, fund, latest.ID)
}

func (a *Analyzer) readDocumentText(doc *models.FundDocument) (string, error) {
	if doc.ExtractedText != "" {
		return doc.ExtractedText, nil
	}
	if doc.SourceURL != "" {
		return "URL: " + doc.SourceURL, nil
	}
	return "", nil
}

func (a *Analyzer) extractMetrics(ctx context.Context, llmClient *Client, docText string) (*MetricsExtraction, error) {
	resp, err := llmClient.ChatSimple(ctx, ExtractMetricsPrompt, docText)
	if err != nil {
		return nil, err
	}
	cleaned := extractJSONObject(resp)
	if cleaned == "" {
		return &MetricsExtraction{}, nil
	}
	var metrics MetricsExtraction
	if err := json.Unmarshal([]byte(cleaned), &metrics); err != nil {
		return &MetricsExtraction{}, nil
	}
	return &metrics, nil
}

func (a *Analyzer) generateAnalysis(ctx context.Context, llmClient *Client, docText string, fund *models.Fund) (*AnalysisResult, error) {
	contextText := fmt.Sprintf("Фонд: %s\nISIN: %s\nУК: %s\n\nДокумент:\n%s", fund.Name, fund.ISIN, fund.ManagementCompany, docText)
	resp, err := llmClient.ChatSimple(ctx, AnalyzePrompt, contextText)
	if err != nil {
		return nil, err
	}
	cleaned := extractJSONObject(resp)
	if cleaned == "" {
		return &AnalysisResult{
			Summary:        truncateString(resp, 1000),
			RiskAssessment: "не удалось получить",
		}, nil
	}
	var result AnalysisResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return &AnalysisResult{
			Summary:        truncateString(resp, 1000),
			RiskAssessment: "не удалось получить",
		}, nil
	}
	return &result, nil
}

func (a *Analyzer) updateFinancialsFromMetrics(fundID uint, metrics *MetricsExtraction) error {
	if metrics == nil {
		return nil
	}

	latest, _ := a.financialsRepo.GetLatestByFundID(fundID)
	if latest == nil {
		latest = &models.FundFinancials{}
	}

	latest.FundID = fundID
	latest.SnapshotDate = time.Now()

	if metrics.UnitPriceRub != nil {
		latest.UnitPriceRub = *metrics.UnitPriceRub
	}
	if metrics.NavPerUnitRub != nil {
		latest.NavPerUnitRub = *metrics.NavPerUnitRub
	}
	if metrics.NavTotalMlnRub != nil {
		latest.NavTotalMlnRub = *metrics.NavTotalMlnRub
	}
	if metrics.DiscountToNavPct != nil {
		latest.DiscountToNavPct = *metrics.DiscountToNavPct
	}
	if metrics.CapRatePct != nil {
		latest.CapRatePct = *metrics.CapRatePct
	}
	if metrics.PNav != nil {
		latest.PNav = *metrics.PNav
	}
	if metrics.PAFFO != nil {
		latest.PAFFO = *metrics.PAFFO
	}
	if metrics.NoiYieldPct != nil {
		latest.NoiYieldPct = *metrics.NoiYieldPct
	}
	if metrics.AnnualPayoutRub != nil {
		latest.AnnualPayoutRub = *metrics.AnnualPayoutRub
	}
	if metrics.PayoutYieldPct != nil {
		latest.PayoutYieldPct = *metrics.PayoutYieldPct
	}
	if metrics.PayoutFrequency != "" {
		latest.PayoutFrequency = metrics.PayoutFrequency
	}
	if metrics.ManagementFeePct != nil {
		latest.ManagementFeePct = *metrics.ManagementFeePct
	}
	if metrics.TradingVolumeMlnRub != nil {
		latest.TradingVolumeMlnRub = *metrics.TradingVolumeMlnRub
	}
	if metrics.NumberOfProperties != nil {
		latest.NumberOfProperties = *metrics.NumberOfProperties
	}

	latest.ID = 0
	return a.financialsRepo.Create(latest)
}

func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(s, "}")
	if end == -1 || end <= start {
		return ""
	}
	return s[start : end+1]
}

func IsPDF(data []byte) bool {
	return len(data) > 4 && string(data[:4]) == "%PDF"
}

func ExtractTextFromPDF(data []byte) (string, error) {
	return "PDF content placeholder\n" + truncateString(string(data[:min(1000, len(data))]), 500), nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}