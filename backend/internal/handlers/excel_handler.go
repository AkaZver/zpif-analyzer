package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zpif-analyzer/backend/internal/services"
)

type ExcelHandler struct {
	excelService *services.ExcelService
}

func NewExcelHandler(excelService *services.ExcelService) *ExcelHandler {
	return &ExcelHandler{excelService: excelService}
}

// ExportExcel godoc
func (h *ExcelHandler) ExportExcel(c *gin.Context) {
	data, err := h.excelService.ExportToExcel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=zpif-export.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
