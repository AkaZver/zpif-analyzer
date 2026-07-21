package repositories

import (
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

type AnalysisRepository struct {
	db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

func (r *AnalysisRepository) GetByFundID(fundID uint) ([]models.LLMAnalysis, error) {
	var analyses []models.LLMAnalysis
	err := r.db.Where("fund_id = ?", fundID).Order("created_at DESC").Find(&analyses).Error
	return analyses, err
}

func (r *AnalysisRepository) GetLatestByFundID(fundID uint) (*models.LLMAnalysis, error) {
	var analysis models.LLMAnalysis
	err := r.db.Where("fund_id = ?", fundID).Order("created_at DESC").First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *AnalysisRepository) Create(analysis *models.LLMAnalysis) error {
	return r.db.Create(analysis).Error
}

func (r *AnalysisRepository) GetByID(id uint) (*models.LLMAnalysis, error) {
	var analysis models.LLMAnalysis
	err := r.db.First(&analysis, id).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}
