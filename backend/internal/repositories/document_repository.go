package repositories

import (
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) GetByFundID(fundID uint) ([]models.FundDocument, error) {
	var documents []models.FundDocument
	err := r.db.Where("fund_id = ?", fundID).Order("upload_date DESC").Find(&documents).Error
	return documents, err
}

func (r *DocumentRepository) GetByID(id uint) (*models.FundDocument, error) {
	var document models.FundDocument
	err := r.db.First(&document, id).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) GetByHash(hash string) (*models.FundDocument, error) {
	var document models.FundDocument
	err := r.db.Where("content_hash = ?", hash).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) Create(document *models.FundDocument) error {
	return r.db.Create(document).Error
}

func (r *DocumentRepository) Update(document *models.FundDocument) error {
	return r.db.Save(document).Error
}

func (r *DocumentRepository) Delete(id uint) error {
	return r.db.Delete(&models.FundDocument{}, id).Error
}

func (r *DocumentRepository) GetPendingByFundID(fundID uint) ([]models.FundDocument, error) {
	var documents []models.FundDocument
	err := r.db.Where("fund_id = ? AND status = ?", fundID, "pending").Find(&documents).Error
	return documents, err
}

func (r *DocumentRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.FundDocument{}).Where("id = ?", id).Update("status", status).Error
}
