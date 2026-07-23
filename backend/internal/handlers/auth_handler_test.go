package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/zpif-analyzer/backend/internal/config"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"github.com/zpif-analyzer/backend/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestAuthHandler(t *testing.T) (*AuthHandler, *gin.Engine, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	userRepo := repositories.NewUserRepository(gormDB)
	authService := services.NewAuthService(userRepo)

	cfg := &config.Config{JWTSecret: "test-secret"}
	handler := NewAuthHandler(authService, cfg)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return handler, router, mock, cleanup
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	_, router, _, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	handler, _, _, _ := setupTestAuthHandler(t)
	router.POST("/api/auth/login", handler.Login)

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, router, mock, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// Mock successful authentication
	now := time.Now()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", string(hashedPassword), "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	router.POST("/api/auth/login", handler.Login)

	body := `{"username":"admin","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "admin", response.User["username"])
}

func TestAuthHandler_GetMe_Success(t *testing.T) {
	handler, router, mock, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// First, login to get a valid token
	now := time.Now()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	loginRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", string(hashedPassword), "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("admin", 1).
		WillReturnRows(loginRows)

	router.POST("/api/auth/login", handler.Login)

	body := `{"username":"admin","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)
	token := loginResponse.Token

	// Now test GetMe with the valid token
	userRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "username", "password_hash", "email", "is_active",
	}).AddRow(1, now, now, nil, "admin", "hash", "admin@test.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(uint(1), 1).
		WillReturnRows(userRows)

	router.GET("/api/auth/me", handler.AuthMiddleware(), handler.GetMe)

	req2 := httptest.NewRequest("GET", "/api/auth/me", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var user map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "admin", user["username"])
	assert.Equal(t, "admin@test.com", user["email"])
}

func TestAuthHandler_GetMe_Unauthorized(t *testing.T) {
	handler, router, _, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	router.GET("/api/auth/me", handler.AuthMiddleware(), handler.GetMe)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetMe_InvalidToken(t *testing.T) {
	handler, router, _, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	router.GET("/api/auth/me", handler.AuthMiddleware(), handler.GetMe)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_AuthMiddleware_NoHeader(t *testing.T) {
	handler, router, _, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	router.GET("/test", handler.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_AuthMiddleware_ValidToken(t *testing.T) {
	handler, router, mock, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// First, login to get a token
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "username", "password_hash", "email", "is_active", "created_at", "updated_at",
	}).AddRow(1, "admin", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", "admin@test.com", true, now, now)

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1`).
		WithArgs("admin").
		WillReturnRows(rows)

	router.POST("/api/auth/login", handler.Login)

	body := `{"username":"admin","password":"admin"}`
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		token := response["token"].(string)

		// Now test GetMe with the token
		rows2 := sqlmock.NewRows([]string{
			"id", "username", "password_hash", "email", "is_active", "created_at", "updated_at",
		}).AddRow(1, "admin", "hash", "admin@test.com", true, now, now)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1`).
			WithArgs(uint(1)).
			WillReturnRows(rows2)

		router.GET("/api/auth/me", handler.AuthMiddleware(), handler.GetMe)

		req2 := httptest.NewRequest("GET", "/api/auth/me", nil)
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
	}
}

func TestLLMHandler_GetSettings(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	defer func() { sqlDB, _ := gormDB.DB(); sqlDB.Close() }()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	llmService := services.NewLLMService(settingsRepo)
	handler := NewLLMHandler(llmService)

	mock.ExpectQuery(`SELECT \* FROM "llm_settings"`).
		WillReturnError(gorm.ErrRecordNotFound)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/llm/settings", handler.GetSettings)

	req := httptest.NewRequest("GET", "/api/llm/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLLMHandler_TestConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	defer func() { sqlDB, _ := gormDB.DB(); sqlDB.Close() }()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	llmService := services.NewLLMService(settingsRepo)
	handler := NewLLMHandler(llmService)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "api_key_encrypted", "base_url", "model_name",
	}).AddRow(1, now, now, nil, "test-api-key", "https://api.openai.com/v1", "gpt-4o-mini")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "llm_settings" WHERE "llm_settings"."deleted_at" IS NULL ORDER BY "llm_settings"."id" LIMIT $1`)).
		WithArgs(1).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/llm/test", handler.TestConnection)

	req := httptest.NewRequest("POST", "/api/llm/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Will fail because we can't connect to real OpenAI, but that's expected
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExcelHandler_ExportExcel(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	defer func() { sqlDB, _ := gormDB.DB(); sqlDB.Close() }()

	fundRepo := repositories.NewFundRepository(gormDB)
	financialsRepo := repositories.NewFinancialsRepository(gormDB)
	analysisRepo := repositories.NewAnalysisRepository(gormDB)
	excelService := services.NewExcelService(fundRepo, financialsRepo, analysisRepo)
	handler := NewExcelHandler(excelService)

	// Mock empty funds list
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "isin", "ticker", "management_company", "real_estate_segment", "qualified_required", "has_market_maker", "fund_end_date", "created_at", "updated_at", "deleted_at"}))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/export/excel", handler.ExportExcel)

	req := httptest.NewRequest("GET", "/api/export/excel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLLMHandler_UpdateSettings(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	defer func() { sqlDB, _ := gormDB.DB(); sqlDB.Close() }()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	llmService := services.NewLLMService(settingsRepo)
	handler := NewLLMHandler(llmService)

	mock.ExpectQuery(`SELECT \* FROM "llm_settings"`).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "llm_settings"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/api/llm/settings", handler.UpdateSettings)

	body := `{"api_key_encrypted":"key","base_url":"https://api.openai.com/v1","model_name":"gpt-4"}`
	req := httptest.NewRequest("PUT", "/api/llm/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLLMHandler_UpdateSettings_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	defer func() { sqlDB, _ := gormDB.DB(); sqlDB.Close() }()

	settingsRepo := repositories.NewLLMSettingsRepository(gormDB)
	llmService := services.NewLLMService(settingsRepo)
	handler := NewLLMHandler(llmService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/api/llm/settings", handler.UpdateSettings)

	req := httptest.NewRequest("PUT", "/api/llm/settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_GetAllFunds_ServiceError(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "funds" WHERE "funds"."deleted_at" IS NULL`)).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/api/funds", handler.GetAllFunds)

	req := httptest.NewRequest("GET", "/api/funds", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFundHandler_UpdateFund_InvalidJSON(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.PUT("/api/funds/:id", handler.UpdateFund)

	req := httptest.NewRequest("PUT", "/api/funds/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFundHandler_DeleteFund_ServiceError(t *testing.T) {
	handler, router, mock, cleanup := setupTestFundHandler(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "funds" SET "deleted_at"=`).
		WithArgs(sqlmock.AnyArg(), uint(1)).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	router.DELETE("/api/funds/:id", handler.DeleteFund)

	req := httptest.NewRequest("DELETE", "/api/funds/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFundHandler_AddFinancials_InvalidJSON(t *testing.T) {
	handler, router, _, cleanup := setupTestFundHandler(t)
	defer cleanup()

	router.POST("/api/funds/:id/financials", handler.AddFinancials)

	req := httptest.NewRequest("POST", "/api/funds/1/financials", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
