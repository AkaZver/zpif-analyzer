package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
)

func TestNewDiscoverer(t *testing.T) {
	llmClient := NewClient("test-key", "", "gpt-4o-mini")
	documentRepo := nil
	fundRepo := nil

	discoverer := NewDiscoverer(llmClient, documentRepo, fundRepo)

	assert.NotNil(t, discoverer)
	assert.Equal(t, llmClient, discoverer.llmClient)
}

func TestDiscoverer_GetStatus_Idle(t *testing.T) {
	llmClient := NewClient("test-key", "", "gpt-4o-mini")
	discoverer := NewDiscoverer(llmClient, nil, nil)

	status := discoverer.GetStatus(123)

	assert.NotNil(t, status)
	assert.Equal(t, uint(123), status.FundID)
	assert.Equal(t, "idle", status.Status)
}

func TestBuildDiscoveryPrompt(t *testing.T) {
	fund := &models.Fund{
		Name:              "Парус ОЗН",
		ISIN:              "RU000A1022Z1",
		ManagementCompany: "Парус Управление Активами",
		RealEstateSegment: "склады",
		Ticker:            "PARUS",
	}

	prompt := buildDiscoveryPrompt(fund)

	assert.Contains(t, prompt, "Парус ОЗН")
	assert.Contains(t, prompt, "RU000A1022Z1")
	assert.Contains(t, prompt, "Парус Управление Активами")
	assert.Contains(t, prompt, "склады")
	assert.Contains(t, prompt, "PARUS")
}

func TestBuildDiscoveryPrompt_Minimal(t *testing.T) {
	fund := &models.Fund{
		Name:              "Тест",
		ISIN:              "RU000TEST001",
		ManagementCompany: "Тест УК",
	}

	prompt := buildDiscoveryPrompt(fund)

	assert.Contains(t, prompt, "Тест")
	assert.Contains(t, prompt, "RU000TEST001")
	assert.NotContains(t, prompt, "Сегмент")
	assert.NotContains(t, prompt, "Тикер")
}
