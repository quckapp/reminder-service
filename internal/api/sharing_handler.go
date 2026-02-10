package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type SharingHandler struct {
	service *service.SharingService
}

func NewSharingHandler(svc *service.SharingService) *SharingHandler {
	return &SharingHandler{service: svc}
}

func (h *SharingHandler) ShareReminder(c *gin.Context) {
	reminderID := c.Param("id")
	sharedBy := c.Query("user_id")

	var req models.ShareReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share, err := h.service.Share(c.Request.Context(), reminderID, sharedBy, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": share})
}

func (h *SharingHandler) GetSharesByReminder(c *gin.Context) {
	reminderID := c.Param("id")
	shares, err := h.service.GetSharesByReminder(c.Request.Context(), reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": shares})
}

func (h *SharingHandler) GetSharedWithUser(c *gin.Context) {
	userID := c.Param("user_id")
	shares, err := h.service.GetSharedWithUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": shares})
}

func (h *SharingHandler) UnshareReminder(c *gin.Context) {
	reminderID := c.Param("id")
	sharedWith := c.Query("shared_with")
	if sharedWith == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shared_with is required"})
		return
	}

	if err := h.service.Unshare(c.Request.Context(), reminderID, sharedWith); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
