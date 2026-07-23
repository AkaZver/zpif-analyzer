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
	api.GET("/funds/:id", fundHandler.GetFundByID)
	api.POST("/funds", fundHandler.CreateFund)
	api.POST("/funds/enrich-and-create", fundHandler.EnrichAndCreateFund)
	api.PUT("/funds/:id", fundHandler.UpdateFund)
	api.DELETE("/funds/:id", fundHandler.DeleteFund)

	// Fund financials
	api.GET("/funds/:id/financials", fundHandler.GetFinancialsByFundID)
	api.POST("/funds/:id/financials", fundHandler.AddFinancials)

	// Fund documents
	api.GET("/funds/:id/documents", fundHandler.GetDocumentsByFundID)
	api.POST("/funds/:id/documents", fundHandler.UploadDocument)
	api.DELETE("/funds/:id/documents/:docId", fundHandler.DeleteDocument)
	api.GET("/funds/:id/documents/:docId/download", fundHandler.DownloadDocument)
	api.POST("/funds/:id/discover", fundHandler.DiscoverDocuments)
	api.GET("/funds/:id/discovery-status", fundHandler.GetDiscoveryStatus)

	// Fund analysis
	api.GET("/funds/:id/analysis", fundHandler.GetLatestAnalysis)
	api.POST("/funds/:id/analyze", fundHandler.AnalyzeFund)

	// Discover all
	api.POST("/funds/discover-all", fundHandler.DiscoverAllDocuments)

	// Market data
	api.POST("/funds/:id/fetch-market-data", marketDataHandler.FetchMarketData)
	api.POST("/funds/fetch-all-market-data", marketDataHandler.FetchAllMarketData)

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
