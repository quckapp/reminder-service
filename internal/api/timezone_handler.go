package api

import (
	"net/http"
	"strings"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type TimezoneHandler struct {
	svc *service.TimezoneService
}

func NewTimezoneHandler(svc *service.TimezoneService) *TimezoneHandler {
	return &TimezoneHandler{svc: svc}
}

func (h *TimezoneHandler) GetUserTimezone(c *gin.Context) {
	userID := c.Param("user_id")

	tz, err := h.svc.GetUserTimezone(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": tz})
}

func (h *TimezoneHandler) SetUserTimezone(c *gin.Context) {
	userID := c.Param("user_id")

	var req models.UpdateTimezoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tz, err := h.svc.SetUserTimezone(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": tz})
}

func (h *TimezoneHandler) ConvertTime(c *gin.Context) {
	var req models.TimezoneConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.ConvertTime(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *TimezoneHandler) GetWorldClock(c *gin.Context) {
	tzStr := c.Query("timezones")
	if tzStr == "" {
		tzStr = "UTC,America/New_York,Europe/London,Asia/Tokyo"
	}
	timezones := strings.Split(tzStr, ",")

	entries, err := h.svc.GetWorldClock(c.Request.Context(), timezones)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": entries})
}

func (h *TimezoneHandler) ListTimezones(c *gin.Context) {
	timezones := h.svc.ListAvailableTimezones(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"success": true, "data": timezones})
}
