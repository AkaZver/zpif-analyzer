package repositories

import (
	"fmt"

	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

type FundRepository struct {
	db *gorm.DB
}

func NewFundRepository(db *gorm.DB) *FundRepository {
	return &FundRepository{db: db}
}

func (r *FundRepository) GetAll() ([]models.Fund, error) {
	var funds []models.Fund
	err := r.db.Preload("Financials").Preload("Documents").Preload("Analyses").Find(&funds).Error
	return funds, err
}

func (r *FundRepository) GetByID(id uint) (*models.Fund, error) {
	var fund models.Fund
	err := r.db.Preload("Financials").Preload("Documents").Preload("Analyses").First(&fund, id).Error
	if err != nil {
		return nil, err
	}
	return &fund, nil
}

func (r *FundRepository) GetByISIN(isin string) (*models.Fund, error) {
	var fund models.Fund
	err := r.db.Where("isin = ?", isin).First(&fund).Error
	if err != nil {
		return nil, err
	}
	return &fund, nil
}

func (r *FundRepository) Create(fund *models.Fund) error {
	return r.db.Create(fund).Error
}

func (r *FundRepository) Update(fund *models.Fund) error {
	return r.db.Save(fund).Error
}

func (r *FundRepository) Delete(id uint) error {
	// Каскадное удаление связанных записей
	if err := r.db.Where("fund_id = ?", id).Delete(&models.FundFinancials{}).Error; err != nil {
		return fmt.Errorf("failed to delete financials: %w", err)
	}

	if err := r.db.Where("fund_id = ?", id).Delete(&models.FundDocument{}).Error; err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	if err := r.db.Where("fund_id = ?", id).Delete(&models.LLMAnalysis{}).Error; err != nil {
		return fmt.Errorf("failed to delete analyses: %w", err)
	}

	// Hard delete фонда
	return r.db.Unscoped().Delete(&models.Fund{}, id).Error
}
