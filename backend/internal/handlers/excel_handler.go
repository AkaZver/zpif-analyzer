package handlers

import (
	"fmt"
	"io"
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

// ImportExcel godoc
func (h *ExcelHandler) ImportExcel(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	
	imported, err := h.excelService.ImportFromExcel(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"imported": imported,
		"message":  fmt.Sprintf("Successfully imported %d records", imported),
	})
}
