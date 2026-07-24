package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/zpif-analyzer/backend/internal/llm"
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
		return &models.LLMSettings{
			BaseURL:   "https://api.openai.com/v1",
			ModelName: "gpt-4o-mini",
		}, nil
	}
	return settings, nil
}

func (s *LLMService) UpdateSettings(settings *models.LLMSettings) error {
	if strings.Contains(settings.APIKeyEncrypted, "****") {
		existing, err := s.settingsRepo.Get()
		if err == nil {
			settings.APIKeyEncrypted = existing.APIKeyEncrypted
		}
	}
	if strings.Contains(settings.ProxyPassword, "****") {
		existing, err := s.settingsRepo.Get()
		if err == nil {
			settings.ProxyPassword = existing.ProxyPassword
		}
	}
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

	proxy := &llm.ProxyConfig{
		Enabled:  settings.ProxyEnabled,
		URL:      settings.ProxyURL,
		Username: settings.ProxyUsername,
		Password: settings.ProxyPassword,
	}
	client := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.ModelName, proxy)
	ctx := context.Background()
	return client.TestConnection(ctx)
}

func (s *LLMService) ListModels() ([]string, error) {
	settings, err := s.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	if settings.APIKeyEncrypted == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	proxy := &llm.ProxyConfig{
		Enabled:  settings.ProxyEnabled,
		URL:      settings.ProxyURL,
		Username: settings.ProxyUsername,
		Password: settings.ProxyPassword,
	}
	client := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.ModelName, proxy)
	ctx := context.Background()
	return client.ListModels(ctx)
}
