package repositories

import (
	"github.com/zpif-analyzer/backend/internal/models"
	"gorm.io/gorm"
)

type LLMSettingsRepository struct {
	db *gorm.DB
}

func NewLLMSettingsRepository(db *gorm.DB) *LLMSettingsRepository {
	return &LLMSettingsRepository{db: db}
}

func (r *LLMSettingsRepository) Get() (*models.LLMSettings, error) {
	var settings models.LLMSettings
	err := r.db.First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *LLMSettingsRepository) Upsert(settings *models.LLMSettings) error {
	var existing models.LLMSettings
	err := r.db.First(&existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new
		return r.db.Create(settings).Error
	} else if err != nil {
		return err
	}
	
	// Update existing
	existing.APIKeyEncrypted = settings.APIKeyEncrypted
	existing.BaseURL = settings.BaseURL
	existing.ModelName = settings.ModelName
	return r.db.Save(&existing).Error
}
