package handlers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zpif-analyzer/backend/internal/llm"
	"github.com/zpif-analyzer/backend/internal/models"
	"github.com/zpif-analyzer/backend/internal/services"
)

type FundHandler struct {
	fundService *services.FundService
}

func NewFundHandler(fundService *services.FundService) *FundHandler {
	return &FundHandler{fundService: fundService}
}

// GetAllFunds godoc
func (h *FundHandler) GetAllFunds(c *gin.Context) {
	funds, err := h.fundService.GetAllFunds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funds)
}

// GetFundByID godoc
func (h *FundHandler) GetFundByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	fund, err := h.fundService.GetFundByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fund not found"})
		return
	}
	
	c.JSON(http.StatusOK, fund)
}

// CreateFund godoc
func (h *FundHandler) CreateFund(c *gin.Context) {
	var fund models.Fund
	if err := c.ShouldBindJSON(&fund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.fundService.CreateFund(&fund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, fund)
}

// UpdateFund godoc
func (h *FundHandler) UpdateFund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	var fund models.Fund
	if err := c.ShouldBindJSON(&fund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.fundService.UpdateFund(uint(id), &fund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, fund)
}

// DeleteFund godoc
func (h *FundHandler) DeleteFund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	if err := h.fundService.DeleteFund(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "fund deleted"})
}

// GetFinancialsByFundID godoc
func (h *FundHandler) GetFinancialsByFundID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	financials, err := h.fundService.GetFinancialsByFundID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, financials)
}

// AddFinancials godoc
func (h *FundHandler) AddFinancials(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	var financials models.FundFinancials
	if err := c.ShouldBindJSON(&financials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.fundService.AddFinancials(uint(id), &financials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, financials)
}

// GetDocumentsByFundID godoc
func (h *FundHandler) GetDocumentsByFundID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	documents, err := h.fundService.GetDocumentsByFundID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, documents)
}

// DeleteDocument godoc
func (h *FundHandler) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}
	
	if err := h.fundService.DeleteDocument(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "document deleted"})
}

// DownloadDocument godoc
func (h *FundHandler) DownloadDocument(c *gin.Context) {
	docID, err := strconv.ParseUint(c.Param("docId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}
	
	document, err := h.fundService.GetDocumentByID(uint(docID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", document.FileName))
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(document.ExtractedText))
}

// GetLatestAnalysis godoc
func (h *FundHandler) GetLatestAnalysis(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}
	
	analysis, err := h.fundService.GetLatestAnalysis(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}
	
	c.JSON(http.StatusOK, analysis)
}

// DiscoverDocuments godoc
func (h *FundHandler) DiscoverDocuments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}

	if err := h.fundService.DiscoverDocumentsForFund(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document discovery completed"})
}

// AnalyzeFund godoc
func (h *FundHandler) AnalyzeFund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}

	analysis, err := h.fundService.AnalyzeFund(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, analysis)
}

// UploadDocument godoc
func (h *FundHandler) UploadDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}

	file, header, err := c.Request.FormFile("file")
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

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	var extractedText string
	if llm.IsPDF(data) {
		extractedText, err = llm.ExtractTextFromPDF(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to extract text from PDF"})
			return
		}
	} else {
		extractedText = string(data)
	}

	document := &models.FundDocument{
		FundID:        uint(id),
		FileName:      header.Filename,
		DocumentType:  c.PostForm("document_type"),
		ContentHash:   hash,
		Source:        "manual",
		UploadDate:    time.Now(),
		Status:        "downloaded",
		FileSize:      int64(len(data)),
		ExtractedText: extractedText,
	}

	if err := h.fundService.AddDocument(document); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, document)
}

// DiscoverAllDocuments godoc
func (h *FundHandler) DiscoverAllDocuments(c *gin.Context) {
	if err := h.fundService.DiscoverDocumentsForAllFunds(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "document discovery started for all funds"})
}

// GetDiscoveryStatus godoc
func (h *FundHandler) GetDiscoveryStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fund ID"})
		return
	}

	status := h.fundService.GetDiscoveryStatus(uint(id))
	c.JSON(http.StatusOK, status)
}

// EnrichAndCreateFund godoc
func (h *FundHandler) EnrichAndCreateFund(c *gin.Context) {
	var req struct {
		Input string `json:"input" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input is required"})
		return
	}

	fund, err := h.fundService.EnrichAndCreateFund(c.Request.Context(), req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fund)
}
