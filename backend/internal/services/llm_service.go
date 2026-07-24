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
			BaseURL:           "https://api.openai.com/v1",
			SearchModelName:   "gpt-4o-mini",
			AnalysisModelName: "gpt-4o-mini",
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

type ModelTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type TestConnectionResult struct {
	SearchModel   ModelTestResult `json:"search_model"`
	AnalysisModel ModelTestResult `json:"analysis_model"`
}

func (s *LLMService) TestConnection() (*TestConnectionResult, error) {
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

	result := &TestConnectionResult{}

	if settings.SearchModelName == settings.AnalysisModelName {
		client := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.SearchModelName, proxy)
		ctx := context.Background()
		err := client.TestConnection(ctx)
		
		if err != nil {
			result.SearchModel = ModelTestResult{Success: false, Message: err.Error()}
			result.AnalysisModel = ModelTestResult{Success: false, Message: err.Error()}
		} else {
			result.SearchModel = ModelTestResult{Success: true, Message: "Connection successful"}
			result.AnalysisModel = ModelTestResult{Success: true, Message: "Connection successful"}
		}
	} else {
		searchClient := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.SearchModelName, proxy)
		ctx := context.Background()
		err := searchClient.TestConnection(ctx)
		if err != nil {
			result.SearchModel = ModelTestResult{Success: false, Message: err.Error()}
		} else {
			result.SearchModel = ModelTestResult{Success: true, Message: "Connection successful"}
		}

		analysisClient := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.AnalysisModelName, proxy)
		err = analysisClient.TestConnection(ctx)
		if err != nil {
			result.AnalysisModel = ModelTestResult{Success: false, Message: err.Error()}
		} else {
			result.AnalysisModel = ModelTestResult{Success: true, Message: "Connection successful"}
		}
	}

	return result, nil
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
	client := llm.NewClient(settings.APIKeyEncrypted, settings.BaseURL, settings.SearchModelName, proxy)
	ctx := context.Background()
	return client.ListModels(ctx)
}
