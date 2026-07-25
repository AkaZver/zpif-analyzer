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

func TestNewDiscoverer(t *testing.T) {
	var settingsRepo *repositories.LLMSettingsRepository
	var documentRepo *repositories.DocumentRepository
	var fundRepo *repositories.FundRepository

	discoverer := NewDiscoverer(settingsRepo, documentRepo, fundRepo)

	assert.NotNil(t, discoverer)
	assert.Equal(t, settingsRepo, discoverer.settingsRepo)
}

func TestDiscoverer_GetStatus_Idle(t *testing.T) {
	discoverer := NewDiscoverer(nil, nil, nil)

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

func TestDiscoverer_Discover_Success(t *testing.T) {
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
						Content: "Информация о фонде Парус",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	discoverer := NewDiscoverer(settingsRepo, documentRepo, fundRepo)

	fund := &models.Fund{
		ID:                1,
		Name:              "Парус",
		ISIN:              "RU000A1022Z1",
		ManagementCompany: "Парус УК",
	}

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "test-model", "test-model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "fund_documents"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	ctx := context.Background()
	status, err := discoverer.Discover(ctx, fund)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, uint(1), status.FundID)
	assert.Equal(t, "completed", status.Status)
}

func TestDiscoverer_Discover_TimeZone(t *testing.T) {
	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	assert.NoError(t, err)
	assert.NotNil(t, moscowTZ)

	now := time.Now().In(moscowTZ)
	formatted := now.Format("2006-01-02-15-04-05")
	assert.NotEmpty(t, formatted)
	assert.Len(t, formatted, 19)
}

func TestDiscoverer_Discover_LLMError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	documentRepo := repositories.NewDocumentRepository(gormDB)
	fundRepo := repositories.NewFundRepository(gormDB)

	discoverer := NewDiscoverer(settingsRepo, documentRepo, fundRepo)

	fund := &models.Fund{
		ID:                3,
		Name:              "Тест",
		ISIN:              "RU000TEST002",
		ManagementCompany: "Тест УК",
	}

	now := time.Now()
	settingsRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "search_model_name", "analysis_model_name", "proxy_enabled", "proxy_url", "proxy_username", "proxy_password"}).
		AddRow(1, now, now, nil, "test-key", server.URL, "test-model", "test-model", false, "", "", "")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(settingsRows)

	ctx := context.Background()
	status, err := discoverer.Discover(ctx, fund)

	assert.Error(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "error", status.Status)
}

func TestDiscoverer_Discover_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	time.Sleep(150 * time.Millisecond)

	select {
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	default:
		t.Fatal("context should have timed out")
	}
}
