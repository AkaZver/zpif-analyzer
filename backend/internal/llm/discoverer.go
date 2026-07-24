package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

type Discoverer struct {
	settingsRepo *repositories.LLMSettingsRepository
	documentRepo *repositories.DocumentRepository
	fundRepo     *repositories.FundRepository
}

func NewDiscoverer(
	settingsRepo *repositories.LLMSettingsRepository,
	documentRepo *repositories.DocumentRepository,
	fundRepo *repositories.FundRepository,
) *Discoverer {
	return &Discoverer{
		settingsRepo: settingsRepo,
		documentRepo: documentRepo,
		fundRepo:     fundRepo,
	}
}

type DiscoveryStatus struct {
	FundID     uint
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
}

var discoveryStorage sync.Map

func (d *Discoverer) Discover(ctx context.Context, fund *models.Fund) (*DiscoveryStatus, error) {
	status := &DiscoveryStatus{
		FundID:    fund.ID,
		Status:    "in_progress",
		StartedAt: time.Now(),
	}
	discoveryStorage.Store(fund.ID, status)
	defer func() {
		status.FinishedAt = time.Now()
		if status.Status == "in_progress" {
			status.Status = "completed"
		}
		discoveryStorage.Store(fund.ID, status)
	}()

	settings, err := d.settingsRepo.Get()
	if err != nil {
		status.Status = "error"
		return status, fmt.Errorf("failed to get LLM settings: %w", err)
	}

	proxy := &ProxyConfig{
		Enabled:  settings.ProxyEnabled,
		URL:      settings.ProxyURL,
		Username: settings.ProxyUsername,
		Password: settings.ProxyPassword,
	}
	llmClient := NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.ModelName, proxy)

	prompt := buildDiscoveryPrompt(fund)

	response, err := llmClient.ChatSimple(ctx, InternetSearchPrompt, prompt)
	if err != nil {
		status.Status = "error"
		return status, fmt.Errorf("LLM search failed: %w", err)
	}

	document := &models.FundDocument{
		FundID:        fund.ID,
		FileName:      fmt.Sprintf("llm-search-%s.txt", time.Now().Format("2006-01-02-15-04-05")),
		DocumentType:  "search",
		ContentHash:   fmt.Sprintf("llm-%s-%d", fund.ISIN, time.Now().Unix()),
		Source:        "auto",
		UploadDate:    time.Now(),
		Status:        "downloaded",
		FileSize:      int64(len(response)),
		ExtractedText: response,
	}

	if err := d.documentRepo.Create(document); err != nil {
		status.Status = "error"
		return status, fmt.Errorf("failed to save document: %w", err)
	}

	status.Status = "completed"
	return status, nil
}

func (d *Discoverer) GetStatus(fundID uint) *DiscoveryStatus {
	val, ok := discoveryStorage.Load(fundID)
	if !ok {
		return &DiscoveryStatus{FundID: fundID, Status: "idle"}
	}
	return val.(*DiscoveryStatus)
}

func buildDiscoveryPrompt(fund *models.Fund) string {
	prompt := fmt.Sprintf("Фонд: %s\nISIN: %s\nУправляющая компания: %s", fund.Name, fund.ISIN, fund.ManagementCompany)
	if fund.RealEstateSegment != "" {
		prompt += fmt.Sprintf("\nСегмент недвижимости: %s", fund.RealEstateSegment)
	}
	if fund.Ticker != "" {
		prompt += fmt.Sprintf("\nТикер: %s", fund.Ticker)
	}
	return prompt
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
