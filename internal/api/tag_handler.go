package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	service *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{service: svc}
}

func (h *TagHandler) CreateTag(c *gin.Context) {
	userID := c.Query("user_id")
	workspaceID := c.Query("workspace_id")

	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.service.Create(c.Request.Context(), userID, workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": tag})
}

func (h *TagHandler) GetTag(c *gin.Context) {
	id := c.Param("id")
	tag, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tag})
}

func (h *TagHandler) GetUserTags(c *gin.Context) {
	userID := c.Param("user_id")
	workspaceID := c.Query("workspace_id")

	tags, err := h.service.GetByUser(c.Request.Context(), userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tags})
}

func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tag})
}

func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TagHandler) TagReminder(c *gin.Context) {
	reminderID := c.Param("id")
	var req models.TagReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.TagReminder(c.Request.Context(), reminderID, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TagHandler) UntagReminder(c *gin.Context) {
	reminderID := c.Param("id")
	tagID := c.Param("tag_id")

	if err := h.service.UntagReminder(c.Request.Context(), reminderID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TagHandler) GetReminderTags(c *gin.Context) {
	reminderID := c.Param("id")
	tags, err := h.service.GetReminderTags(c.Request.Context(), reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tags})
}

func (h *TagHandler) GetRemindersByTag(c *gin.Context) {
	tagID := c.Param("id")
	ids, err := h.service.GetRemindersByTag(c.Request.Context(), tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ids})
}

func (h *TagHandler) BulkTag(c *gin.Context) {
	var req models.BulkTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.BulkTag(c.Request.Context(), req.ReminderIDs, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
