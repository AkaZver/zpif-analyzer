package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zpif-analyzer/backend/internal/config"
	"github.com/zpif-analyzer/backend/internal/handlers"
	"github.com/zpif-analyzer/backend/internal/llm"
	"github.com/zpif-analyzer/backend/internal/middleware"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/parsers"
	"github.com/zpif-analyzer/backend/internal/repositories"
	"github.com/zpif-analyzer/backend/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Загрузка конфигурации
	cfg := config.Load()

	// Формирование DSN для PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	// Подключение к базе данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Автоматическая миграция схем
	err = db.AutoMigrate(
		&models.Fund{},
		&models.FundFinancials{},
		&models.FundDocument{},
		&models.LLMAnalysis{},
		&models.LLMSettings{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed")

	// Миграция LLM настроек: копирование model_name в новые поля
	migrateLLMSettings(db)

	// Создание начальных данных (seed)
	seedInitialData(db, cfg)

	// Инициализация repositories
	fundRepo := repositories.NewFundRepository(db)
	financialsRepo := repositories.NewFinancialsRepository(db)
	documentRepo := repositories.NewDocumentRepository(db)
	analysisRepo := repositories.NewAnalysisRepository(db)
	llmSettingsRepo := repositories.NewLLMSettingsRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Инициализация services
	fundService := services.NewFundService(fundRepo, financialsRepo, documentRepo, analysisRepo)
	fundService.SetLLMSettingsRepo(llmSettingsRepo)
	authService := services.NewAuthService(userRepo)
	llmService := services.NewLLMService(llmSettingsRepo)
	excelService := services.NewExcelService(fundRepo, financialsRepo, analysisRepo)

	// Инициализация парсеров рыночных данных
	moexParser := parsers.NewMoexParser()
	investfundsParser := parsers.NewInvestfundsParser()
	vsezpifParser := parsers.NewVsezpifParser()
	marketDataService := services.NewMarketDataService(moexParser, investfundsParser, vsezpifParser, financialsRepo, fundRepo)
	log.Println("Market data parsers initialized")

	// Инициализация LLM компонентов (настройки берутся из БД при каждом вызове)
	discoverer := llm.NewDiscoverer(llmSettingsRepo, documentRepo, fundRepo)
	fundService.SetDiscoverer(discoverer)
	analyzer := llm.NewAnalyzer(llmSettingsRepo, documentRepo, analysisRepo, financialsRepo, fundRepo)
	fundService.SetAnalyzer(analyzer)
	log.Println("LLM components initialized")

	// Инициализация handlers
	fundHandler := handlers.NewFundHandler(fundService)
	authHandler := handlers.NewAuthHandler(authService, cfg)
	llmHandler := handlers.NewLLMHandler(llmService)
	excelHandler := handlers.NewExcelHandler(excelService)
	marketDataHandler := handlers.NewMarketDataHandler(marketDataService)

	// Настройка Gin router
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	r.POST("/api/auth/login", authHandler.Login)

	// Protected routes
	api := r.Group("/api")
	api.Use(authHandler.AuthMiddleware())

	// Funds
	api.GET("/funds", fundHandler.GetAllFunds)
	api.GET("/funds-with-financials", fundHandler.GetAllFundsWithLatestFinancials)
	api.POST("/funds", fundHandler.CreateFund)
	api.POST("/funds/enrich-and-create", fundHandler.EnrichAndCreateFund)
	api.POST("/funds/discover-all", fundHandler.DiscoverAllDocuments)
	api.POST("/funds/fetch-all-market-data", marketDataHandler.FetchAllMarketData)

	fundByID := api.Group("/funds/:id")
	fundByID.GET("", fundHandler.GetFundByID)
	fundByID.PUT("", fundHandler.UpdateFund)
	fundByID.DELETE("", fundHandler.DeleteFund)
	fundByID.GET("/financials", fundHandler.GetFinancialsByFundID)
	fundByID.POST("/financials", fundHandler.AddFinancials)
	fundByID.GET("/documents", fundHandler.GetDocumentsByFundID)
	fundByID.POST("/documents", fundHandler.UploadDocument)
	fundByID.DELETE("/documents/:docId", fundHandler.DeleteDocument)
	fundByID.GET("/documents/:docId/download", fundHandler.DownloadDocument)
	fundByID.POST("/discover", fundHandler.DiscoverDocuments)
	fundByID.GET("/discovery-status", fundHandler.GetDiscoveryStatus)
	fundByID.GET("/analysis", fundHandler.GetLatestAnalysis)
	fundByID.POST("/analyze", fundHandler.AnalyzeFund)
	fundByID.POST("/fetch-market-data", marketDataHandler.FetchMarketData)

	// Auth
	api.GET("/auth/me", authHandler.GetMe)

	// LLM settings
	api.GET("/llm/settings", llmHandler.GetSettings)
	api.PUT("/llm/settings", llmHandler.UpdateSettings)
	api.POST("/llm/test", llmHandler.TestConnection)
	api.GET("/llm/models", llmHandler.ListModels)

	// Excel export
	api.GET("/export/excel", excelHandler.ExportExcel)

	// Запуск сервера
	port := cfg.ServerPort
	log.Printf("Server starting on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func seedInitialData(db *gorm.DB, cfg *config.Config) {
	// Проверка, есть ли уже данные
	var count int64
	db.Model(&models.Fund{}).Count(&count)

	if count > 0 {
		log.Println("Initial data already exists, skipping seed")
		return
	}

	log.Println("Seeding initial data...")

	// Начальные фонды из плана
	funds := []models.Fund{
		{
			Name:              "Парус ОЗН",
			ISIN:              "RU000A1022Z1",
			ManagementCompany: "Парус Управление Активами",
		},
		{
			Name:              "Акцент 5",
			ISIN:              "RU000A10DQF7",
			ManagementCompany: "Акцент",
		},
		{
			Name:              "ВИМ РД",
			ISIN:              "RU000A102N77",
			ManagementCompany: "ВИМ",
		},
		{
			Name:              "Современная коллекция",
			ISIN:              "RU000A10CQ02",
			ManagementCompany: "Сбер",
		},
	}

	for _, fund := range funds {
		if err := db.Create(&fund).Error; err != nil {
			log.Printf("Failed to seed fund %s: %v", fund.Name, err)
		} else {
			log.Printf("Seeded fund: %s", fund.Name)
		}
	}

	// Начальный пользователь (admin)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	user := models.User{
		Username:     "admin",
		PasswordHash: string(hashedPassword),
		Email:        "admin@zpif-analyzer.local",
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Printf("Failed to seed admin user: %v", err)
	} else {
		log.Println("Seeded admin user (username: admin)")
	}

	log.Println("Initial data seeded successfully")
}

func migrateLLMSettings(db *gorm.DB) {
	// Проверяем наличие старого поля model_name
	if !db.Migrator().HasColumn(&models.LLMSettings{}, "model_name") {
		return
	}

	// Получаем все записи с model_name
	var settings []models.LLMSettings
	if err := db.Find(&settings).Error; err != nil {
		log.Printf("Failed to fetch LLM settings for migration: %v", err)
		return
	}

	// Копируем model_name в новые поля если они пустые
	for _, s := range settings {
		needsUpdate := false
		if s.SearchModelName == "" {
			db.Model(&s).Update("search_model_name", db.Raw("SELECT model_name FROM llm_settings WHERE id = ?", s.ID))
			needsUpdate = true
		}
		if s.AnalysisModelName == "" {
			db.Model(&s).Update("analysis_model_name", db.Raw("SELECT model_name FROM llm_settings WHERE id = ?", s.ID))
			needsUpdate = true
		}
		if needsUpdate {
			log.Printf("Migrated LLM settings: model_name -> search_model_name, analysis_model_name")
		}
	}

	// Удаляем старое поле
	db.Migrator().DropColumn(&models.LLMSettings{}, "model_name")
	log.Println("Removed deprecated model_name column")
}
