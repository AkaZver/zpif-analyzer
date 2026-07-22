package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zpif-analyzer/backend/internal/services"
)

type MarketDataHandler struct {
	marketDataService *services.MarketDataService
}

func NewMarketDataHandler(service *services.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{marketDataService: service}
}

func (h *MarketDataHandler) FetchMarketData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}

	result, err := h.marketDataService.FetchMarketDataForFund(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MarketDataHandler) FetchAllMarketData(c *gin.Context) {
	result, err := h.marketDataService.FetchMarketDataForAllFunds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
