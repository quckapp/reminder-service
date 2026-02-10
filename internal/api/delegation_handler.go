package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type DelegationHandler struct {
	service *service.DelegationService
}

func NewDelegationHandler(svc *service.DelegationService) *DelegationHandler {
	return &DelegationHandler{service: svc}
}

func (h *DelegationHandler) Delegate(c *gin.Context) {
	reminderID := c.Param("id")
	delegatedBy := c.Query("user_id")

	var req models.DelegateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delegation, err := h.service.Delegate(c.Request.Context(), reminderID, delegatedBy, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": delegation})
}

func (h *DelegationHandler) Accept(c *gin.Context) {
	delegationID := c.Param("id")

	delegation, err := h.service.Accept(c.Request.Context(), delegationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": delegation})
}

func (h *DelegationHandler) Reject(c *gin.Context) {
	delegationID := c.Param("id")

	delegation, err := h.service.Reject(c.Request.Context(), delegationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": delegation})
}

func (h *DelegationHandler) GetDelegatedReminders(c *gin.Context) {
	userID := c.Param("user_id")

	delegations, err := h.service.GetDelegatedReminders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": delegations})
}
