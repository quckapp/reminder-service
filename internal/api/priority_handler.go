package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PriorityHandler struct {
	service *service.PriorityService
}

func NewPriorityHandler(svc *service.PriorityService) *PriorityHandler {
	return &PriorityHandler{service: svc}
}

func (h *PriorityHandler) SetPriority(c *gin.Context) {
	id := c.Param("id")
	var req models.SetPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SetPriority(c.Request.Context(), id, req.Priority); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *PriorityHandler) ListByPriority(c *gin.Context) {
	userID := c.Param("user_id")
	priority := models.ReminderPriority(c.Query("priority"))

	reminders, err := h.service.ListByPriority(c.Request.Context(), userID, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminders})
}

func (h *PriorityHandler) GetDistribution(c *gin.Context) {
	userID := c.Param("user_id")

	dist, err := h.service.GetDistribution(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": dist})
}
