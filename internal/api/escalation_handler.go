package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type EscalationHandler struct {
	service *service.EscalationService
}

func NewEscalationHandler(svc *service.EscalationService) *EscalationHandler {
	return &EscalationHandler{service: svc}
}

func (h *EscalationHandler) CreateRule(c *gin.Context) {
	userID := c.Query("user_id")
	workspaceID := c.Query("workspace_id")

	var req models.CreateEscalationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.service.CreateRule(c.Request.Context(), userID, workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": rule})
}

func (h *EscalationHandler) ListRules(c *gin.Context) {
	userID := c.Query("user_id")
	workspaceID := c.Query("workspace_id")

	rules, err := h.service.ListRules(c.Request.Context(), userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rules})
}

func (h *EscalationHandler) UpdateRule(c *gin.Context) {
	ruleID := c.Param("id")

	var req models.UpdateEscalationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.service.UpdateRule(c.Request.Context(), ruleID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rule})
}

func (h *EscalationHandler) DeleteRule(c *gin.Context) {
	ruleID := c.Param("id")

	if err := h.service.DeleteRule(c.Request.Context(), ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *EscalationHandler) GetHistory(c *gin.Context) {
	reminderID := c.Param("id")

	events, err := h.service.GetHistory(c.Request.Context(), reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
}
