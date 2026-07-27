package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func TestExtractJSON_ValidObject(t *testing.T) {
	input := `Some text before {"key":"value"} and after`
	result := ExtractJSON(input)

	assert.Equal(t, `{"key":"value"}`, result)
}

func TestExtractJSON_NoObject(t *testing.T) {
	input := `No JSON object here`
	result := ExtractJSON(input)

	assert.Equal(t, input, result)
}

func TestExtractJSON_IncompleteObject(t *testing.T) {
	input := `{"key":"value"`
	result := ExtractJSON(input)

	assert.Equal(t, input, result)
}

func TestExtractJSON_MarkdownCodeBlock(t *testing.T) {
	input := "```json\n{\"key\":\"value\"}\n```"
	result := ExtractJSON(input)

	assert.Equal(t, `{"key":"value"}`, result)
}

func TestExtractJSON_MarkdownCodeBlockWithText(t *testing.T) {
	input := "Вот результат:\n```json\n{\"name\":\"Фонд\"}\n```\nГотово!"
	result := ExtractJSON(input)

	assert.Equal(t, `{"name":"Фонд"}`, result)
}

func TestExtractJSON_RussianTextAround(t *testing.T) {
	input := `Вот данные фонда: {"name":"ЗПИФ Склады","isin":"RU000A1022Z1"} Это всё.`
	result := ExtractJSON(input)

	assert.Equal(t, `{"name":"ЗПИФ Склады","isin":"RU000A1022Z1"}`, result)
}

func TestExtractJSON_SmartQuotes(t *testing.T) {
	input := "{\"name\":\"ЗПИФ «ПАРУС»\",\"value\":\"\u201ctest\u201d\"}"
	result := ExtractJSON(input)

	assert.Contains(t, result, "«ПАРУС»")
	assert.Contains(t, result, "\"test\"")
	assert.NotContains(t, result, "\u201c")
	assert.NotContains(t, result, "\u201d")
}

func TestExtractJSON_BOM(t *testing.T) {
	input := "\xef\xbb\xbf{\"key\":\"value\"}"
	result := ExtractJSON(input)

	assert.Equal(t, `{"key":"value"}`, result)
}

func TestExtractJSON_NestedObjects(t *testing.T) {
	input := `{"outer":{"inner":"value"},"array":[1,2,3]}`
	result := ExtractJSON(input)

	assert.Equal(t, input, result)
}

func TestExtractJSON_EscapedQuotes(t *testing.T) {
	input := `{"text":"He said \"hello\" to me"}`
	result := ExtractJSON(input)

	assert.Equal(t, input, result)
}

func TestExtractJSON_SmartDoubleQuotesAsDelimiters(t *testing.T) {
	input := "{\"name\": \"ЗПИФ\", \"isin\": \"RU000A1022Z1\"}"
	result := ExtractJSON(input)

	assert.Equal(t, `{"name": "ЗПИФ", "isin": "RU000A1022Z1"}`, result)
}

func TestExtractJSON_SmartDoubleQuotesInsideString(t *testing.T) {
	input := `{"name":"ЗПИФ «ПАРУС»"}`
	result := ExtractJSON(input)

	assert.Contains(t, result, "«ПАРУС»")
}

func TestExtractJSON_SmartSingleQuotesInsideString(t *testing.T) {
	input := "{\"name\":\"it\u2019s a test\"}"
	result := ExtractJSON(input)

	assert.Contains(t, result, "it's a test")
}

func TestSanitizeJSON_UnescapedQuotes(t *testing.T) {
	input := `{"name":"ЗПИФ недвижимости "ПАРУС-Двинцев""}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"name":"ЗПИФ недвижимости 'ПАРУС-Двинцев'"}`, result)
}

func TestSanitizeJSON_ValidJSON(t *testing.T) {
	input := `{"name":"Test","value":123}`
	result := SanitizeJSON(input)

	assert.Equal(t, input, result)
}

func TestSanitizeJSON_EscapedQuotes(t *testing.T) {
	input := `{"text":"He said \"hello\""}`
	result := SanitizeJSON(input)

	assert.Equal(t, input, result)
}

func TestSanitizeJSON_MultipleUnescapedQuotes(t *testing.T) {
	input := `{"name":"Фонд "Альфа" и "Бета""}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"name":"Фонд 'Альфа' и 'Бета'"}`, result)
}

func TestSanitizeJSON_SingleQuotesAsDelimiters(t *testing.T) {
	input := `{'name': 'ЗПИФ', 'isin': 'RU000A1022Z1'}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"name": "ЗПИФ", "isin": "RU000A1022Z1"}`, result)
}

func TestSanitizeJSON_SingleQuotesInsideDoubleQuotedString(t *testing.T) {
	input := `{"name": "it's a test"}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"name": "it's a test"}`, result)
}

func TestSanitizeJSON_MixedQuotes(t *testing.T) {
	input := `{'name': "ЗПИФ"}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"name": "ЗПИФ"}`, result)
}

func TestSanitizeJSON_EmptyValues(t *testing.T) {
	input := `{"name": "", "value": ""}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"name": "", "value": ""}`, result)
}

func TestSanitizeJSON_Arrays(t *testing.T) {
	input := `{"items": ["a", "b", "c"]}`
	result := SanitizeJSON(input)

	assert.Equal(t, `{"items": ["a", "b", "c"]}`, result)
}

func TestExtractAndSanitize_SmartQuoteDelimiters(t *testing.T) {
	input := "Вот результат:\n```json\n{\"name\": \"ЗПИФ\", \"isin\": \"RU000A1022Z1\"}\n```\nГотово!"
	extracted := ExtractJSON(input)
	result := SanitizeJSON(extracted)

	assert.JSONEq(t, `{"name": "ЗПИФ", "isin": "RU000A1022Z1"}`, result)
}

func TestExtractAndSanitize_SmartQuotesInside(t *testing.T) {
	input := "{\"name\":\"ЗПИФ «ПАРУС»\",\"value\":\"\u201ctest\u201d\"}"
	extracted := ExtractJSON(input)
	result := SanitizeJSON(extracted)

	assert.JSONEq(t, "{\"name\":\"ЗПИФ «ПАРУС»\",\"value\":\"'test'\"}", result)
}

func TestExtractAndSanitize_SingleQuoteJSON(t *testing.T) {
	input := `{'name': 'ЗПИФ', 'isin': 'RU000A1022Z1'}`
	extracted := ExtractJSON(input)
	result := SanitizeJSON(extracted)

	assert.JSONEq(t, `{"name": "ЗПИФ", "isin": "RU000A1022Z1"}`, result)
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

func TestAnalyzer_AnalyzeDocuments_EmptyDocumentIDs(t *testing.T) {
	analyzer := &Analyzer{}
	fund := &models.Fund{ID: 1, Name: "Test Fund"}

	result, err := analyzer.AnalyzeDocuments(context.Background(), fund, []uint{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no documents specified")
}

func TestAnalyzer_AnalyzeDocuments_NilDocumentIDs(t *testing.T) {
	analyzer := &Analyzer{}
	fund := &models.Fund{ID: 1, Name: "Test Fund"}

	result, err := analyzer.AnalyzeDocuments(context.Background(), fund, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no documents specified")
}

func intPtr(i int) *int {
	return &i
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func setupAnalyzerTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestAnalyzer_Analyze_SettingsError(t *testing.T) {
	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund := &models.Fund{ID: 1, Name: "Test Fund"}
	result, err := analyzer.Analyze(context.Background(), fund, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get LLM settings")
}

func TestAnalyzer_Analyze_DocumentNotFound(t *testing.T) {
	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", "http://localhost", "model", "model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	fund := &models.Fund{ID: 1, Name: "Test Fund"}
	result, err := analyzer.Analyze(context.Background(), fund, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "document not found")
}

func TestAnalyzer_Analyze_EmptyDocument(t *testing.T) {
	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", "http://localhost", "model", "model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	docRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status", "file_size", "extracted_text",
	}).AddRow(1, now, now, nil, 1, "empty.pdf", "", "appraisal", "hash", "auto", "", now, "downloaded", 0, "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(docRows)

	fund := &models.Fund{ID: 1, Name: "Test Fund"}
	result, err := analyzer.Analyze(context.Background(), fund, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "document is empty")
}

func TestAnalyzer_Analyze_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		messages := reqBody["messages"].([]interface{})
		systemMsg := messages[0].(map[string]interface{})["content"].(string)

		var content string
		if systemMsg == ExtractMetricsPrompt {
			metrics := MetricsExtraction{
				UnitPriceRub:    floatPtr(1000.0),
				CapRatePct:      floatPtr(8.5),
				PayoutFrequency: "monthly",
			}
			content = toJSON(metrics)
		} else {
			analysis := AnalysisResult{
				Summary:        "Test summary",
				RiskAssessment: "Low risk",
				Pros:           []string{"Good"},
				Cons:           []string{"Bad"},
			}
			content = toJSON(analysis)
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
						Content: content,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	docRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status", "file_size", "extracted_text",
	}).AddRow(1, now, now, nil, 1, "report.pdf", "", "appraisal", "hash", "auto", "", now, "downloaded", 100, "Test document content for analysis")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(docRows)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_analyses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(fund_id = .+ AND snapshot_date <= .+\) AND "fund_financials"\."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"\."id" LIMIT`).
		WithArgs(uint(1), sqlmock.AnyArg(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_documents" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	fund := &models.Fund{ID: 1, Name: "Test Fund", ISIN: "RU000TEST001", ManagementCompany: "Test UK"}
	result, err := analyzer.Analyze(context.Background(), fund, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test summary", result.AnalysisSummary)
}

func TestAnalyzer_AnalyzeLatestDocuments_NoDocs(t *testing.T) {
	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY upload_date DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	fund := &models.Fund{ID: 1, Name: "Test Fund"}
	result, err := analyzer.AnalyzeLatestDocuments(context.Background(), fund)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no documents to analyze")
}

func TestAnalyzer_AnalyzeLatestDocuments_AllAnalyzed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		messages := reqBody["messages"].([]interface{})
		systemMsg := messages[0].(map[string]interface{})["content"].(string)

		var content string
		if systemMsg == ExtractMetricsPrompt {
			content = toJSON(MetricsExtraction{})
		} else {
			content = toJSON(AnalysisResult{Summary: "Summary", RiskAssessment: "Low"})
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
						Content: content,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	now := time.Now()

	docRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status", "file_size", "extracted_text",
	}).AddRow(1, now, now, nil, 1, "analyzed.pdf", "", "appraisal", "hash", "auto", "", now, "analyzed", 100, "Analyzed document content")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE fund_id = $1 AND "fund_documents"."deleted_at" IS NULL ORDER BY upload_date DESC`)).
		WithArgs(uint(1)).
		WillReturnRows(docRows)

	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	docByIDRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status", "file_size", "extracted_text",
	}).AddRow(1, now, now, nil, 1, "analyzed.pdf", "", "appraisal", "hash", "auto", "", now, "analyzed", 100, "Analyzed document content")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(docByIDRows)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_analyses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(fund_id = .+ AND snapshot_date <= .+\) AND "fund_financials"\."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"\."id" LIMIT`).
		WithArgs(uint(1), sqlmock.AnyArg(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_documents" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	fund := &models.Fund{ID: 1, Name: "Test Fund", ISIN: "RU000TEST001", ManagementCompany: "Test UK"}
	result, err := analyzer.AnalyzeLatestDocuments(context.Background(), fund)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAnalyzer_UpdateFinancialsFromMetrics_NilMetrics(t *testing.T) {
	gormDB, _, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	err := analyzer.updateFinancialsFromMetrics(1, nil)

	assert.NoError(t, err)
}

func TestAnalyzer_UpdateFinancialsFromMetrics_NoExisting(t *testing.T) {
	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(fund_id = .+ AND snapshot_date <= .+\) AND "fund_financials"\."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"\."id" LIMIT`).
		WithArgs(uint(1), sqlmock.AnyArg(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	metrics := &MetricsExtraction{
		UnitPriceRub: floatPtr(1000.0),
		CapRatePct:   floatPtr(8.5),
	}

	err := analyzer.updateFinancialsFromMetrics(1, metrics)

	assert.NoError(t, err)
}

func TestAnalyzer_UpdateFinancialsFromMetrics_FullMetrics(t *testing.T) {
	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	now := time.Now()
	snapshotDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	existingRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "snapshot_date",
		"unit_price_rub", "nav_per_unit_rub", "nav_total_mln_rub", "discount_to_nav_pct",
		"cap_rate_pct", "p_nav", "p_affo", "noi_yield_pct",
		"annual_payout_rub", "payout_yield_pct", "payout_yield_after_tax_pct",
		"payout_frequency", "payout_stability", "rent_indexation_pct",
		"management_fee_pct", "trading_volume_mln_rub",
		"number_of_properties", "main_tenants",
	}).AddRow(1, now, now, nil, 1, snapshotDate, 900.0, 950.0, 4000.0, -5.0,
		8.0, 0.9, 11.0, 7.0, 70.0, 7.0, 6.0, "quarterly", "medium", 2.5, 1.0, 4.0, 2, "Old Tenant")

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(fund_id = .+ AND snapshot_date <= .+\) AND "fund_financials"\."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"\."id" LIMIT`).
		WithArgs(uint(1), sqlmock.AnyArg(), 1).
		WillReturnRows(existingRows)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	metrics := &MetricsExtraction{
		UnitPriceRub:        floatPtr(1000.0),
		NavPerUnitRub:       floatPtr(1050.0),
		NavTotalMlnRub:      floatPtr(5000.0),
		DiscountToNavPct:    floatPtr(-4.76),
		CapRatePct:          floatPtr(8.5),
		PNav:                floatPtr(0.95),
		PAFFO:               floatPtr(12.0),
		NoiYieldPct:         floatPtr(7.2),
		AnnualPayoutRub:     floatPtr(80.0),
		PayoutYieldPct:      floatPtr(8.0),
		PayoutFrequency:     "monthly",
		ManagementFeePct:    floatPtr(1.5),
		TradingVolumeMlnRub: floatPtr(5.0),
		NumberOfProperties:  intPtr(3),
	}

	err := analyzer.updateFinancialsFromMetrics(1, metrics)

	assert.NoError(t, err)
}

func TestExtractTextFromPDF(t *testing.T) {
	data := []byte("%PDF-1.4 test content here")
	text, err := ExtractTextFromPDF(data)

	assert.NoError(t, err)
	assert.Contains(t, text, "PDF content placeholder")
}

func TestAnalyzer_AnalyzeDocuments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		messages := reqBody["messages"].([]interface{})
		systemMsg := messages[0].(map[string]interface{})["content"].(string)

		var content string
		if systemMsg == ExtractMetricsPrompt {
			metrics := MetricsExtraction{
				UnitPriceRub: floatPtr(1000.0),
			}
			content = toJSON(metrics)
		} else {
			analysis := AnalysisResult{
				Summary:        "Multi-doc summary",
				RiskAssessment: "Medium risk",
				Pros:           []string{"Pro 1"},
				Cons:           []string{"Con 1"},
			}
			content = toJSON(analysis)
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
						Content: content,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	gormDB, mock, cleanup := setupAnalyzerTestDB(t)
	defer cleanup()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	analyzer := NewAnalyzer(settingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "model", "model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	docRows1 := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status", "file_size", "extracted_text",
	}).AddRow(1, now, now, nil, 1, "doc1.pdf", "", "appraisal", "hash1", "auto", "", now, "downloaded", 100, "Document 1 content")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(1), 1).
		WillReturnRows(docRows1)

	docRows2 := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "fund_id", "file_name", "file_path", "document_type",
		"content_hash", "source", "source_url", "upload_date", "status", "file_size", "extracted_text",
	}).AddRow(2, now, now, nil, 1, "doc2.pdf", "", "report", "hash2", "auto", "", now, "downloaded", 200, "Document 2 content")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fund_documents" WHERE "fund_documents"."id" =`)).
		WithArgs(uint(2), 1).
		WillReturnRows(docRows2)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_analyses"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT \* FROM "fund_financials" WHERE \(fund_id = .+ AND snapshot_date <= .+\) AND "fund_financials"\."deleted_at" IS NULL ORDER BY snapshot_date DESC,"fund_financials"\."id" LIMIT`).
		WithArgs(uint(1), sqlmock.AnyArg(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "fund_financials"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_documents" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "fund_documents" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	fund := &models.Fund{ID: 1, Name: "Test Fund", ISIN: "RU000TEST001", ManagementCompany: "Test UK"}
	result, err := analyzer.AnalyzeDocuments(context.Background(), fund, []uint{1, 2})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Multi-doc summary", result.AnalysisSummary)
}

func TestAnalyzer_ExtractMetrics_NoJSON(t *testing.T) {
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
						Content: "no json here at all",
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
}

func TestAnalyzer_GenerateAnalysis_ParseError(t *testing.T) {
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
						Content: `{"summary": invalid json here}`,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llmClient := NewClient("test-key", server.URL, "gpt-4o-mini", nil)
	analyzer := &Analyzer{}

	fund := &models.Fund{ID: 1, Name: "Test Fund"}
	result, err := analyzer.generateAnalysis(context.Background(), llmClient, "test document", fund)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "не удалось получить", result.RiskAssessment)
}
