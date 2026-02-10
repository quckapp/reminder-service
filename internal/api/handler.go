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

func RegisterRoutes(
	router *gin.Engine,
	svc *service.ReminderService,
	tagHandler *TagHandler,
	templateHandler *TemplateHandler,
	notifHandler *NotificationHandler,
	sharingHandler *SharingHandler,
	noteHandler *NoteHandler,
	analyticsHandler *AnalyticsHandler,
	priorityHandler *PriorityHandler,
	subtaskHandler *SubtaskHandler,
	calendarHandler *CalendarHandler,
	escalationHandler *EscalationHandler,
	categoryHandler *CategoryHandler,
	delegationHandler *DelegationHandler,
	recurringHandler *RecurringHandler,
	timezoneHandler *TimezoneHandler,
	habitHandler *HabitHandler,
) {
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
		api.POST("/reminders/:id/complete", h.CompleteReminder)

		api.POST("/reminders/bulk", h.BulkCreateReminders)
		api.POST("/reminders/bulk-cancel", h.BulkCancelReminders)
		api.POST("/reminders/bulk-delete", h.BulkDeleteReminders)
		api.POST("/reminders/bulk-snooze", analyticsHandler.BulkSnooze)
		api.POST("/reminders/bulk-complete", analyticsHandler.BulkComplete)

		api.POST("/reminders/:id/tags", tagHandler.TagReminder)
		api.GET("/reminders/:id/tags", tagHandler.GetReminderTags)
		api.DELETE("/reminders/:id/tags/:tag_id", tagHandler.UntagReminder)
		api.POST("/tags", tagHandler.CreateTag)
		api.GET("/tags/:id", tagHandler.GetTag)
		api.PUT("/tags/:id", tagHandler.UpdateTag)
		api.DELETE("/tags/:id", tagHandler.DeleteTag)
		api.GET("/tags/:id/reminders", tagHandler.GetRemindersByTag)
		api.POST("/tags/bulk", tagHandler.BulkTag)

		api.POST("/reminders/:id/notes", noteHandler.CreateNote)
		api.GET("/reminders/:id/notes", noteHandler.GetNotes)
		api.PUT("/reminders/notes/:note_id", noteHandler.UpdateNote)
		api.DELETE("/reminders/notes/:note_id", noteHandler.DeleteNote)

		api.POST("/reminders/:id/share", sharingHandler.ShareReminder)
		api.GET("/reminders/:id/shares", sharingHandler.GetSharesByReminder)
		api.DELETE("/reminders/:id/share", sharingHandler.UnshareReminder)

		api.GET("/reminders/:id/activity", analyticsHandler.GetReminderActivity)

		api.POST("/templates", templateHandler.CreateTemplate)
		api.GET("/templates/:id", templateHandler.GetTemplate)
		api.PUT("/templates/:id", templateHandler.UpdateTemplate)
		api.DELETE("/templates/:id", templateHandler.DeleteTemplate)
		api.POST("/templates/create-reminder", templateHandler.CreateFromTemplate)
		api.GET("/templates/popular", templateHandler.GetPopularTemplates)

		api.GET("/users/:user_id/notification-preferences", notifHandler.GetPreferences)
		api.PUT("/users/:user_id/notification-preferences", notifHandler.UpdatePreferences)
		api.DELETE("/users/:user_id/notification-preferences", notifHandler.DeletePreferences)

		api.GET("/users/:user_id/reminders", h.GetUserReminders)
		api.GET("/users/:user_id/reminders/stats", h.GetUserReminderStats)
		api.GET("/users/:user_id/reminders/upcoming", analyticsHandler.GetUpcoming)
		api.GET("/users/:user_id/reminders/overdue", analyticsHandler.GetOverdue)
		api.GET("/users/:user_id/reminders/today", analyticsHandler.GetDueToday)
		api.GET("/users/:user_id/templates", templateHandler.GetUserTemplates)
		api.GET("/users/:user_id/tags", tagHandler.GetUserTags)
		api.GET("/users/:user_id/shared", sharingHandler.GetSharedWithUser)
		api.GET("/users/:user_id/activity", analyticsHandler.GetUserActivity)
		api.GET("/channels/:channel_id/reminders", h.GetChannelReminders)
		api.GET("/workspaces/:workspace_id/reminders", h.GetWorkspaceReminders)
		api.GET("/workspaces/:workspace_id/analytics", analyticsHandler.GetWorkspaceAnalytics)

		api.GET("/search", analyticsHandler.SearchReminders)
		api.GET("/export", analyticsHandler.ExportReminders)
		api.POST("/import", analyticsHandler.ImportReminders)

		// -- Priority --
		api.PUT("/reminders/:id/priority", priorityHandler.SetPriority)
		api.GET("/users/:user_id/reminders/by-priority", priorityHandler.ListByPriority)
		api.GET("/users/:user_id/reminders/priority-distribution", priorityHandler.GetDistribution)

		// -- Subtasks --
		api.POST("/reminders/:id/subtasks", subtaskHandler.AddSubtask)
		api.GET("/reminders/:id/subtasks", subtaskHandler.ListSubtasks)
		api.PUT("/subtasks/:subtask_id", subtaskHandler.UpdateSubtask)
		api.DELETE("/subtasks/:subtask_id", subtaskHandler.DeleteSubtask)
		api.POST("/subtasks/:subtask_id/toggle", subtaskHandler.ToggleSubtask)

		// -- Calendar Integration --
		api.GET("/calendar/export", calendarHandler.ExportICal)
		api.POST("/calendar/feed", calendarHandler.GetFeedURL)
		api.POST("/calendar/sync", calendarHandler.SyncCalendar)
		api.GET("/calendar/view", calendarHandler.GetCalendarView)

		// -- Escalation Rules --
		api.POST("/escalation-rules", escalationHandler.CreateRule)
		api.GET("/escalation-rules", escalationHandler.ListRules)
		api.PUT("/escalation-rules/:id", escalationHandler.UpdateRule)
		api.DELETE("/escalation-rules/:id", escalationHandler.DeleteRule)
		api.GET("/reminders/:id/escalation-history", escalationHandler.GetHistory)

		// -- Categories --
		api.POST("/categories", categoryHandler.CreateCategory)
		api.GET("/categories", categoryHandler.ListCategories)
		api.PUT("/categories/:id", categoryHandler.UpdateCategory)
		api.DELETE("/categories/:id", categoryHandler.DeleteCategory)

		// -- Delegation --
		api.POST("/reminders/:id/delegate", delegationHandler.Delegate)
		api.POST("/delegations/:id/accept", delegationHandler.Accept)
		api.POST("/delegations/:id/reject", delegationHandler.Reject)
		api.GET("/users/:user_id/delegated", delegationHandler.GetDelegatedReminders)

		// -- Recurring Patterns --
		api.POST("/recurring-patterns", recurringHandler.CreatePattern)
		api.GET("/recurring-patterns", recurringHandler.ListPatterns)
		api.GET("/recurring-patterns/active", recurringHandler.GetActivePatterns)
		api.GET("/recurring-patterns/:id", recurringHandler.GetPattern)
		api.PUT("/recurring-patterns/:id", recurringHandler.UpdatePattern)
		api.DELETE("/recurring-patterns/:id", recurringHandler.DeletePattern)
		api.POST("/recurring-patterns/:id/toggle", recurringHandler.ToggleActive)
		api.GET("/recurring-patterns/:id/occurrences", recurringHandler.ListOccurrences)

		// -- Timezone --
		api.GET("/users/:user_id/timezone", timezoneHandler.GetUserTimezone)
		api.PUT("/users/:user_id/timezone", timezoneHandler.SetUserTimezone)
		api.POST("/timezone/convert", timezoneHandler.ConvertTime)
		api.GET("/timezone/world-clock", timezoneHandler.GetWorldClock)
		api.GET("/timezone/list", timezoneHandler.ListTimezones)

		// -- Habits --
		api.POST("/habits", habitHandler.CreateHabit)
		api.GET("/habits/:id", habitHandler.GetHabit)
		api.PUT("/habits/:id", habitHandler.UpdateHabit)
		api.DELETE("/habits/:id", habitHandler.DeleteHabit)
		api.POST("/habits/:id/complete", habitHandler.CompleteHabit)
		api.GET("/habits/:id/completions", habitHandler.GetCompletions)
		api.GET("/habits/:id/stats", habitHandler.GetStats)
		api.POST("/habits/:id/reset-streak", habitHandler.ResetStreak)
		api.GET("/users/:user_id/habits", habitHandler.ListHabits)
		api.GET("/users/:user_id/habits/summary", habitHandler.GetSummary)
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

	var page models.PaginationParams
	if err := c.ShouldBindQuery(&page); err != nil {
		page = models.PaginationParams{Page: 1, PerPage: 20}
	}

	result, err := h.service.GetByUserIDPaginated(c.Request.Context(), userID, status, &page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result.Data, "total": result.Total, "page": result.Page, "per_page": result.PerPage, "total_pages": result.TotalPages})
}


func (h *Handler) CompleteReminder(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Complete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) BulkCreateReminders(c *gin.Context) {
	var req models.BulkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.BulkCreate(c.Request.Context(), req.Reminders)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *Handler) BulkCancelReminders(c *gin.Context) {
	var req models.BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.BulkCancel(c.Request.Context(), req.IDs)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *Handler) BulkDeleteReminders(c *gin.Context) {
	var req models.BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.BulkDelete(c.Request.Context(), req.IDs)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *Handler) GetUserReminderStats(c *gin.Context) {
	userID := c.Param("user_id")

	stats, err := h.service.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *Handler) GetChannelReminders(c *gin.Context) {
	channelID := c.Param("channel_id")

	var page models.PaginationParams
	if err := c.ShouldBindQuery(&page); err != nil {
		page = models.PaginationParams{Page: 1, PerPage: 20}
	}

	result, err := h.service.GetByChannelID(c.Request.Context(), channelID, &page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result.Data, "total": result.Total, "page": result.Page, "per_page": result.PerPage, "total_pages": result.TotalPages})
}

func (h *Handler) GetWorkspaceReminders(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	statusStr := c.Query("status")

	var status *models.ReminderStatus
	if statusStr != "" {
		s := models.ReminderStatus(statusStr)
		status = &s
	}

	var page models.PaginationParams
	if err := c.ShouldBindQuery(&page); err != nil {
		page = models.PaginationParams{Page: 1, PerPage: 20}
	}

	result, err := h.service.GetByWorkspaceID(c.Request.Context(), workspaceID, status, &page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result.Data, "total": result.Total, "page": result.Page, "per_page": result.PerPage, "total_pages": result.TotalPages})
}
