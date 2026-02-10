package api

import (
	"net/http"
	"strconv"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	templateSvc *service.TemplateService
	reminderSvc *service.ReminderService
}

func NewTemplateHandler(tmplSvc *service.TemplateService, reminderSvc *service.ReminderService) *TemplateHandler {
	return &TemplateHandler{templateSvc: tmplSvc, reminderSvc: reminderSvc}
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	userID := c.Query("user_id")
	workspaceID := c.Query("workspace_id")

	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.templateSvc.Create(c.Request.Context(), userID, workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": tmpl})
}

func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	tmpl, err := h.templateSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tmpl})
}

func (h *TemplateHandler) GetUserTemplates(c *gin.Context) {
	userID := c.Param("user_id")
	workspaceID := c.Query("workspace_id")

	templates, err := h.templateSvc.GetByUser(c.Request.Context(), userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": templates})
}

func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.templateSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tmpl})
}

func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.templateSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TemplateHandler) CreateFromTemplate(c *gin.Context) {
	var req models.CreateFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.templateSvc.GetByID(c.Request.Context(), req.TemplateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	reminderReq := &models.CreateReminderRequest{
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		ChannelID:   req.ChannelID,
		Type:        tmpl.Type,
		Title:       tmpl.Title,
		Description: tmpl.Description,
		RemindAt:    req.RemindAt,
		Recurrence:  tmpl.Recurrence,
		Metadata:    tmpl.Metadata,
	}

	reminder, err := h.reminderSvc.Create(c.Request.Context(), reminderReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Increment usage count
	_ = h.templateSvc.IncrementUsage(c.Request.Context(), req.TemplateID)

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": reminder})
}

func (h *TemplateHandler) GetPopularTemplates(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	limit := int64(10)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	templates, err := h.templateSvc.GetPopular(c.Request.Context(), workspaceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": templates})
}
