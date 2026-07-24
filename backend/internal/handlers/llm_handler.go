package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/services"
)

type LLMHandler struct {
	llmService *services.LLMService
}

func NewLLMHandler(llmService *services.LLMService) *LLMHandler {
	return &LLMHandler{llmService: llmService}
}

// GetSettings godoc
func (h *LLMHandler) GetSettings(c *gin.Context) {
	settings, err := h.llmService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	proxyPassword := settings.ProxyPassword
	if proxyPassword != "" {
		proxyPassword = "****"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"id":                settings.ID,
		"base_url":          settings.BaseURL,
		"model_name":        settings.ModelName,
		"api_key_encrypted": settings.APIKeyEncrypted,
		"proxy_enabled":     settings.ProxyEnabled,
		"proxy_url":         settings.ProxyURL,
		"proxy_username":    settings.ProxyUsername,
		"proxy_password":    proxyPassword,
		"updated_at":        settings.UpdatedAt,
	})
}

// UpdateSettings godoc
func (h *LLMHandler) UpdateSettings(c *gin.Context) {
	var settings models.LLMSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.llmService.UpdateSettings(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

// TestConnection godoc
func (h *LLMHandler) TestConnection(c *gin.Context) {
	err := h.llmService.TestConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection successful",
	})
}

// ListModels godoc
func (h *LLMHandler) ListModels(c *gin.Context) {
	models, err := h.llmService.ListModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, models)
}
