package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: svc}
}

func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	userID := c.Param("user_id")
	workspaceID := c.Query("workspace_id")

	pref, err := h.service.GetPreferences(c.Request.Context(), userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": pref})
}

func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userID := c.Param("user_id")
	workspaceID := c.Query("workspace_id")

	var req models.UpdateNotificationPrefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pref, err := h.service.UpdatePreferences(c.Request.Context(), userID, workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": pref})
}

func (h *NotificationHandler) DeletePreferences(c *gin.Context) {
	userID := c.Param("user_id")
	workspaceID := c.Query("workspace_id")

	if err := h.service.DeletePreferences(c.Request.Context(), userID, workspaceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
