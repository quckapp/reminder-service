package api

import (
	"net/http"
	"strconv"
	"time"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsSvc *service.AnalyticsService
	searchSvc    *service.SearchService
	activitySvc  *service.ActivityService
	exportSvc    *service.ExportService
	reminderSvc  *service.ReminderService
}

func NewAnalyticsHandler(
	analyticsSvc *service.AnalyticsService,
	searchSvc *service.SearchService,
	activitySvc *service.ActivityService,
	exportSvc *service.ExportService,
	reminderSvc *service.ReminderService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsSvc: analyticsSvc,
		searchSvc:    searchSvc,
		activitySvc:  activitySvc,
		exportSvc:    exportSvc,
		reminderSvc:  reminderSvc,
	}
}

func (h *AnalyticsHandler) GetWorkspaceAnalytics(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	analytics, err := h.analyticsSvc.GetWorkspaceAnalytics(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": analytics})
}

func (h *AnalyticsHandler) GetUpcoming(c *gin.Context) {
	userID := c.Param("user_id")
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	reminders, err := h.analyticsSvc.GetUpcoming(c.Request.Context(), userID, days, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminders})
}

func (h *AnalyticsHandler) GetOverdue(c *gin.Context) {
	userID := c.Param("user_id")
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	reminders, err := h.analyticsSvc.GetOverdue(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminders})
}

func (h *AnalyticsHandler) GetDueToday(c *gin.Context) {
	userID := c.Param("user_id")

	reminders, err := h.analyticsSvc.GetDueToday(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": reminders})
}

// ── Search ──

func (h *AnalyticsHandler) SearchReminders(c *gin.Context) {
	var params models.ReminderSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.searchSvc.Search(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result.Data, "total": result.Total, "page": result.Page, "per_page": result.PerPage, "total_pages": result.TotalPages})
}

// ── Activity ──

func (h *AnalyticsHandler) GetReminderActivity(c *gin.Context) {
	reminderID := c.Param("id")
	limit := int64(50)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	activities, err := h.activitySvc.GetByReminder(c.Request.Context(), reminderID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": activities})
}

func (h *AnalyticsHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("user_id")
	limit := int64(50)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	activities, err := h.activitySvc.GetByUser(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": activities})
}

// ── Export/Import ──

func (h *AnalyticsHandler) ExportReminders(c *gin.Context) {
	userID := c.Query("user_id")
	format := c.Query("format")
	if format == "" {
		format = "json"
	}
	status := c.Query("status")

	req := &models.ExportRequest{
		UserID: userID,
		Format: format,
		Status: status,
	}

	result, err := h.exportSvc.Export(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) ImportReminders(c *gin.Context) {
	userID := c.Query("user_id")
	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.exportSvc.Import(c.Request.Context(), userID, &req)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ── Bulk Extended ──

func (h *AnalyticsHandler) BulkSnooze(c *gin.Context) {
	var req models.BulkSnoozeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format"})
		return
	}

	resp := &models.BulkActionResponse{}
	for _, id := range req.IDs {
		_, err := h.reminderSvc.Snooze(c.Request.Context(), id, duration)
		if err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, err.Error())
		} else {
			resp.Successful++
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *AnalyticsHandler) BulkComplete(c *gin.Context) {
	var req models.BulkCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := &models.BulkActionResponse{}
	for _, id := range req.IDs {
		err := h.reminderSvc.Complete(c.Request.Context(), id)
		if err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, err.Error())
		} else {
			resp.Successful++
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}
