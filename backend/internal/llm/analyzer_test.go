package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
)

func TestNewAnalyzer(t *testing.T) {
	settingsRepo := &repositories.LLMSettingsRepository{}
	documentRepo := &repositories.DocumentRepository{}
	analysisRepo := &repositories.AnalysisRepository{}
	financialsRepo := &repositories.FinancialsRepository{}
	fundRepo := &repositories.FundRepository{}

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	assert.NotNil(t, analyzer)
	assert.Equal(t, settingsRepo, analyzer.settingsRepo)
}

func TestExtractJSONObject_ValidObject(t *testing.T) {
	input := `Some text before {"key":"value"} and after`
	result := extractJSONObject(input)

	assert.Equal(t, `{"key":"value"}`, result)
}

func TestExtractJSONObject_NoObject(t *testing.T) {
	input := `No JSON object here`
	result := extractJSONObject(input)

	assert.Empty(t, result)
}

func TestExtractJSONObject_IncompleteObject(t *testing.T) {
	input := `{"key":"value"`
	result := extractJSONObject(input)

	assert.Empty(t, result)
}

func TestIsPDF_ValidPDF(t *testing.T) {
	data := []byte("%PDF-1.4 test content")
	assert.True(t, IsPDF(data))
}

func TestIsPDF_NotPDF(t *testing.T) {
	data := []byte("HTML content")
	assert.False(t, IsPDF(data))
}

func TestIsPDF_EmptyData(t *testing.T) {
	data := []byte("")
	assert.False(t, IsPDF(data))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 5, min(5, 10))
	assert.Equal(t, 3, min(10, 3))
	assert.Equal(t, 7, min(7, 7))
}

func TestAnalyzer_ReadDocumentText_NoFilePath(t *testing.T) {
	analyzer := &Analyzer{}
	doc := &models.FundDocument{
		FilePath:  "",
		SourceURL: "http://example.com/doc.pdf",
	}

	text, err := analyzer.readDocumentText(doc)

	assert.NoError(t, err)
	assert.Contains(t, text, "http://example.com/doc.pdf")
}

func TestAnalyzer_ReadDocumentText_NoFilePathNoURL(t *testing.T) {
	analyzer := &Analyzer{}
	doc := &models.FundDocument{
		FilePath:  "",
		SourceURL: "",
	}

	text, err := analyzer.readDocumentText(doc)

	assert.NoError(t, err)
	assert.Empty(t, text)
}

func TestAnalyzer_ReadDocumentText_ExtractedText(t *testing.T) {
	analyzer := &Analyzer{}
	doc := &models.FundDocument{
		ExtractedText: "Test content from extracted text",
	}

	text, err := analyzer.readDocumentText(doc)

	assert.NoError(t, err)
	assert.Contains(t, text, "Test content")
}

func TestAnalyzer_ReadDocumentText_EmptyFields(t *testing.T) {
	analyzer := &Analyzer{}
	doc := &models.FundDocument{}

	text, err := analyzer.readDocumentText(doc)

	assert.NoError(t, err)
	assert.Empty(t, text)
}

func TestAnalyzer_ExtractMetrics_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := MetricsExtraction{
			UnitPriceRub:    floatPtr(1000.0),
			CapRatePct:      floatPtr(8.5),
			PayoutFrequency: "monthly",
		}
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: toJSON(metrics),
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llmClient := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	analyzer := &Analyzer{}

	metrics, err := analyzer.extractMetrics(context.Background(), llmClient, "test document")

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.UnitPriceRub)
	assert.Equal(t, 1000.0, *metrics.UnitPriceRub)
	assert.Equal(t, "monthly", metrics.PayoutFrequency)
}

func TestAnalyzer_ExtractMetrics_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "not valid json",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llmClient := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	analyzer := &Analyzer{}

	metrics, err := analyzer.extractMetrics(context.Background(), llmClient, "test document")

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// Should return empty metrics when JSON is invalid
}

func TestAnalyzer_GenerateAnalysis_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		analysis := AnalysisResult{
			Summary:        "Test summary",
			RiskAssessment: "Low risk",
			Pros:           []string{"Good returns", "Stable"},
			Cons:           []string{"High fees"},
		}
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: toJSON(analysis),
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llmClient := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	analyzer := &Analyzer{}

	fund := &models.Fund{
		ID:                1,
		Name:              "Test Fund",
		ISIN:              "RU000TEST001",
		ManagementCompany: "Test UK",
	}

	result, err := analyzer.generateAnalysis(context.Background(), llmClient, "test document", fund)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test summary", result.Summary)
	assert.Equal(t, "Low risk", result.RiskAssessment)
	assert.Len(t, result.Pros, 2)
	assert.Len(t, result.Cons, 1)
}

func TestAnalyzer_GenerateAnalysis_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ChatResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "This is a plain text response without JSON",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llmClient := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	analyzer := &Analyzer{}

	fund := &models.Fund{
		ID:   1,
		Name: "Test Fund",
	}

	result, err := analyzer.generateAnalysis(context.Background(), llmClient, "test document", fund)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should return truncated response when JSON is invalid
	assert.Contains(t, result.Summary, "This is a plain text response")
	assert.Equal(t, "не удалось получить", result.RiskAssessment)
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
