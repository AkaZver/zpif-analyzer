package repositories

import (
	"time"

	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

type FinancialsRepository struct {
	db *gorm.DB
}

func NewFinancialsRepository(db *gorm.DB) *FinancialsRepository {
	return &FinancialsRepository{db: db}
}

func (r *FinancialsRepository) GetByFundID(fundID uint) ([]models.FundFinancials, error) {
	var financials []models.FundFinancials
	err := r.db.Where("fund_id = ?", fundID).Order("snapshot_date DESC").Find(&financials).Error
	return financials, err
}

func (r *FinancialsRepository) GetByFundIDAndDateRange(fundID uint, from, to time.Time) ([]models.FundFinancials, error) {
	var financials []models.FundFinancials
	err := r.db.Where("fund_id = ? AND snapshot_date BETWEEN ? AND ?", fundID, from, to).
		Order("snapshot_date DESC").Find(&financials).Error
	return financials, err
}

func (r *FinancialsRepository) GetLatestByFundID(fundID uint) (*models.FundFinancials, error) {
	var financial models.FundFinancials
	err := r.db.Where("fund_id = ?", fundID).Order("snapshot_date DESC").First(&financial).Error
	if err != nil {
		return nil, err
	}
	return &financial, nil
}

func (r *FinancialsRepository) Create(financial *models.FundFinancials) error {
	return r.db.Create(financial).Error
}

func (r *FinancialsRepository) Update(financial *models.FundFinancials) error {
	return r.db.Save(financial).Error
}

func (r *FinancialsRepository) Delete(id uint) error {
	return r.db.Delete(&models.FundFinancials{}, id).Error
}
