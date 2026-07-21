package services

import (
	"context"
	"fmt"

	"github.com/zpif-analyzer/backend/internal/llm"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"github.com/zpif-analyzer/backend/internal/websearch"
)

type LLMService struct {
	settingsRepo *repositories.LLMSettingsRepository
}

func NewLLMService(settingsRepo *repositories.LLMSettingsRepository) *LLMService {
	return &LLMService{settingsRepo: settingsRepo}
}

func (s *LLMService) GetSettings() (*models.LLMSettings, error) {
	settings, err := s.settingsRepo.Get()
	if err != nil {
		return &models.LLMSettings{
			BaseURL:   "https://api.openai.com/v1",
			ModelName: "gpt-4o-mini",
		}, nil
	}
	return settings, nil
}

func (s *LLMService) UpdateSettings(settings *models.LLMSettings) error {
	return s.settingsRepo.Upsert(settings)
}

func (s *LLMService) TestConnection() error {
	settings, err := s.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}
	if settings.APIKeyEncrypted == "" {
		return fmt.Errorf("API key not configured")
	}

	client := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.ModelName)
	ctx := context.Background()
	return client.TestConnection(ctx)
}

func (s *LLMService) TestWebSearch(provider, apiKey string) (int, error) {
	searchClient, err := websearch.NewClient(provider, apiKey)
	if err != nil {
		return 0, fmt.Errorf("invalid search config: %w", err)
	}

	results, err := searchClient.Search(context.Background(), "ZPIF Парус ОЗН")
	if err != nil {
		return 0, err
	}
	return len(results), nil
}