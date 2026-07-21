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
	
	// Mask API key for security
	maskedSettings := gin.H{
		"id":        settings.ID,
		"base_url":  settings.BaseURL,
		"model":     settings.ModelName,
		"api_key":   maskAPIKey(settings.APIKeyEncrypted),
		"updated_at": settings.UpdatedAt,
	}
	
	c.JSON(http.StatusOK, maskedSettings)
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

// TestWebSearch godoc
func (h *LLMHandler) TestWebSearch(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		APIKey   string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider and api_key are required"})
		return
	}

	results, err := h.llmService.TestWebSearch(req.Provider, req.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
			"results": 0,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Web search test successful",
		"results": results,
	})
}

// Helper function to mask API key
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
