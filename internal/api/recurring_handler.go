package api

import (
	"net/http"
	"strconv"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type RecurringHandler struct {
	svc *service.RecurringService
}

func NewRecurringHandler(svc *service.RecurringService) *RecurringHandler {
	return &RecurringHandler{svc: svc}
}

func (h *RecurringHandler) CreatePattern(c *gin.Context) {
	userID := c.Query("user_id")
	workspaceID := c.Query("workspace_id")

	var req models.CreateRecurringPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pattern, err := h.svc.CreatePattern(c.Request.Context(), userID, workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": pattern})
}

func (h *RecurringHandler) GetPattern(c *gin.Context) {
	id := c.Param("id")

	pattern, err := h.svc.GetPattern(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pattern not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": pattern})
}

func (h *RecurringHandler) ListPatterns(c *gin.Context) {
	userID := c.Query("user_id")

	patterns, err := h.svc.ListPatterns(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": patterns})
}

func (h *RecurringHandler) UpdatePattern(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateRecurringPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pattern, err := h.svc.UpdatePattern(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": pattern})
}

func (h *RecurringHandler) DeletePattern(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeletePattern(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *RecurringHandler) ToggleActive(c *gin.Context) {
	id := c.Param("id")

	pattern, err := h.svc.ToggleActive(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": pattern})
}

func (h *RecurringHandler) ListOccurrences(c *gin.Context) {
	patternID := c.Param("id")
	limit := int64(50)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	occs, err := h.svc.ListOccurrences(c.Request.Context(), patternID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": occs})
}

func (h *RecurringHandler) GetActivePatterns(c *gin.Context) {
	patterns, err := h.svc.GetActivePatterns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": patterns})
}
