package repositories

import (
	"fmt"
	"time"

	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

type FundRepository struct {
	db *gorm.DB
}

func NewFundRepository(db *gorm.DB) *FundRepository {
	return &FundRepository{db: db}
}

const whereFundID = "fund_id = ?"

func (r *FundRepository) GetAll() ([]models.Fund, error) {
	var funds []models.Fund
	err := r.db.Preload("Financials").Preload("Documents").Preload("Analyses").Find(&funds).Error
	return funds, err
}

func (r *FundRepository) GetAllWithLatestFinancials() ([]models.Fund, error) {
	var funds []models.Fund
	err := r.db.Find(&funds).Error
	if err != nil {
		return nil, err
	}

	if len(funds) == 0 {
		return funds, nil
	}

	fundIDs := make([]uint, len(funds))
	for i, f := range funds {
		fundIDs[i] = f.ID
	}

	var latestFinancials []models.FundFinancials
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location())
	err = r.db.
		Where("id IN (SELECT DISTINCT ON (fund_id) id FROM fund_financials WHERE fund_id IN ? AND deleted_at IS NULL AND snapshot_date <= ? ORDER BY fund_id, snapshot_date DESC)", fundIDs, today).
		Find(&latestFinancials).Error
	if err != nil {
		return nil, err
	}

	finMap := make(map[uint]*models.FundFinancials)
	for i := range latestFinancials {
		finMap[latestFinancials[i].FundID] = &latestFinancials[i]
	}

	for i := range funds {
		if fin, ok := finMap[funds[i].ID]; ok {
			funds[i].Financials = []models.FundFinancials{*fin}
		}
	}

	return funds, nil
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
	if err := r.db.Where(whereFundID, id).Delete(&models.FundFinancials{}).Error; err != nil {
		return fmt.Errorf("failed to delete financials: %w", err)
	}

	if err := r.db.Where(whereFundID, id).Delete(&models.FundDocument{}).Error; err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	if err := r.db.Where(whereFundID, id).Delete(&models.LLMAnalysis{}).Error; err != nil {
		return fmt.Errorf("failed to delete analyses: %w", err)
	}

	// Hard delete фонда
	return r.db.Unscoped().Delete(&models.Fund{}, id).Error
}
