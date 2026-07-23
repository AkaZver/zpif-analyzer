package models

import (
	"time"

	"gorm.io/gorm"
)

type Fund struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Основная информация
	Name               string `gorm:"uniqueIndex;not null" json:"name"`
	ISIN               string `gorm:"uniqueIndex;not null" json:"isin"`
	Ticker             string `json:"ticker"`
	ManagementCompany  string `json:"management_company"`
	RealEstateSegment  string `json:"real_estate_segment"`
	QualifiedRequired  bool   `json:"qualified_required"`
	HasMarketMaker     bool   `json:"has_market_maker"`

	// Даты
	FundEndDate *time.Time `json:"fund_end_date"`

	// Внешние источники
	InvestfundsURL string `json:"investfunds_url"`
	VsezpifURL     string `json:"vsezpif_url"`

	// Связи
	Financials []FundFinancials `gorm:"foreignKey:FundID" json:"financials"`
	Documents  []FundDocument   `gorm:"foreignKey:FundID" json:"documents"`
	Analyses   []LLMAnalysis    `gorm:"foreignKey:FundID" json:"analyses"`
}

type FundFinancials struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	FundID       uint      `gorm:"index;not null" json:"fund_id"`
	SnapshotDate time.Time `gorm:"not null" json:"snapshot_date"`

	// Цены и стоимость
	UnitPriceRub     float64 `json:"unit_price_rub"`
	NavPerUnitRub    float64 `json:"nav_per_unit_rub"`
	NavTotalMlnRub   float64 `json:"nav_total_mln_rub"`
	DiscountToNavPct float64 `json:"discount_to_nav_pct"`

	// Ключевые метрики
	CapRatePct   float64 `json:"cap_rate_pct"`
	PNav         float64 `json:"p_nav"`
	PAFFO        float64 `json:"p_affo"`
	NoiYieldPct  float64 `json:"noi_yield_pct"`

	// Выплаты
	AnnualPayoutRub           float64 `json:"annual_payout_rub"`
	PayoutYieldPct            float64 `json:"payout_yield_pct"`
	PayoutYieldAfterTaxPct    float64 `json:"payout_yield_after_tax_pct"`
	PayoutFrequency           string  `json:"payout_frequency"`
	PayoutStability           string  `json:"payout_stability"`
	RentIndexationPct         float64 `json:"rent_indexation_pct"`

	// Операции
	ManagementFeePct       float64 `json:"management_fee_pct"`
	TradingVolumeMlnRub    float64 `json:"trading_volume_mln_rub"`
	NumberOfProperties     int     `json:"number_of_properties"`
	MainTenants            string  `json:"main_tenants"`

	// Связь
	Fund Fund `gorm:"foreignKey:FundID" json:"fund"`
}

type FundDocument struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	FundID         uint   `gorm:"index;not null" json:"fund_id"`
	FileName       string `json:"file_name"`
	FilePath       string `json:"file_path"`
	DocumentType   string `json:"document_type"`
	ContentHash    string `json:"content_hash"`
	Source         string `json:"source"`
	SourceURL      string `json:"source_url"`
	UploadDate     time.Time `json:"upload_date"`
	Status         string    `json:"status"`
	FileSize       int64     `json:"file_size"`
	ExtractedText  string    `gorm:"type:text" json:"extracted_text"`

	// Связь
	Fund Fund `gorm:"foreignKey:FundID" json:"fund"`
}

type LLMAnalysis struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	FundID           uint   `gorm:"index;not null" json:"fund_id"`
	DocumentID       uint   `json:"document_id"`
	ModelUsed        string `json:"model_used"`
	RawResponse      string `gorm:"type:text" json:"raw_response"`
	AnalysisSummary  string `gorm:"type:text" json:"analysis_summary"`
	RiskAssessment   string `gorm:"type:text" json:"risk_assessment"`
	ProsCons         string `gorm:"type:text" json:"pros_cons"`
	ExtractedMetrics string `gorm:"type:jsonb" json:"extracted_metrics"`

	// Связь
	Fund     Fund          `gorm:"foreignKey:FundID" json:"fund"`
	Document *FundDocument `gorm:"foreignKey:DocumentID" json:"document"`
}

type LLMSettings struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	APIKeyEncrypted string `json:"api_key_encrypted"`
	BaseURL         string `json:"base_url"`
	ModelName       string `json:"model_name"`
}

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Username     string `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"not null" json:"-"`
	Email        string `json:"email"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
}
