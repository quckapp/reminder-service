package api

import (
	"net/http"
	"time"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.ReminderService
}

func RegisterRoutes(router *gin.Engine, svc *service.ReminderService) {
	h := &Handler{service: svc}

	// Health endpoints
	router.GET("/health", h.Health)
	router.GET("/health/ready", h.HealthReady)
	router.GET("/health/live", h.HealthLive)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/reminders", h.CreateReminder)
		api.GET("/reminders/:id", h.GetReminder)
		api.PUT("/reminders/:id", h.UpdateReminder)
		api.DELETE("/reminders/:id", h.DeleteReminder)
		api.POST("/reminders/:id/snooze", h.SnoozeReminder)
		api.POST("/reminders/:id/cancel", h.CancelReminder)
		api.GET("/users/:user_id/reminders", h.GetUserReminders)
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "reminder-service",
	})
}

func (h *Handler) HealthReady(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ready": true})
}

func (h *Handler) HealthLive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"live": true})
}

func (h *Handler) CreateReminder(c *gin.Context) {
	var req models.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reminder, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": reminder})
}

func (h *Handler) GetReminder(c *gin.Context) {
	id := c.Param("id")

	reminder, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminder})
}

func (h *Handler) UpdateReminder(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reminder, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminder})
}

func (h *Handler) DeleteReminder(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) SnoozeReminder(c *gin.Context) {
	id := c.Param("id")

	var req models.SnoozeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format"})
		return
	}

	reminder, err := h.service.Snooze(c.Request.Context(), id, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminder})
}

func (h *Handler) CancelReminder(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetUserReminders(c *gin.Context) {
	userID := c.Param("user_id")
	statusStr := c.Query("status")

	var status *models.ReminderStatus
	if statusStr != "" {
		s := models.ReminderStatus(statusStr)
		status = &s
	}

	reminders, err := h.service.GetByUserID(c.Request.Context(), userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminders})
}
