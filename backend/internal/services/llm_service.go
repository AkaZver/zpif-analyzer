package services

import (
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
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
		// Return default settings if not found
		return &models.LLMSettings{
			BaseURL:   "https://api.openai.com/v1",
			ModelName: "gpt-4o-mini",
		}, nil
	}
	return settings, nil
}

func (s *LLMService) UpdateSettings(settings *models.LLMSettings) error {
	// TODO: Implement encryption for API key
	// For now, just store as-is
	return s.settingsRepo.Upsert(settings)
}

func (s *LLMService) TestConnection() error {
	// TODO: Implement actual LLM connection test
	// This will be implemented in Phase 5
	return nil
}

func (s *LLMService) TestWebSearch() (int, error) {
	// TODO: Implement actual web search test
	// This will be implemented in Phase 5
	return 0, nil
}
