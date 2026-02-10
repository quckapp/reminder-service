package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type SubtaskHandler struct {
	service *service.SubtaskService
}

func NewSubtaskHandler(svc *service.SubtaskService) *SubtaskHandler {
	return &SubtaskHandler{service: svc}
}

func (h *SubtaskHandler) AddSubtask(c *gin.Context) {
	reminderID := c.Param("id")
	userID := c.Query("user_id")

	var req models.CreateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtask, err := h.service.Create(c.Request.Context(), reminderID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": subtask})
}

func (h *SubtaskHandler) ListSubtasks(c *gin.Context) {
	reminderID := c.Param("id")

	subtasks, err := h.service.List(c.Request.Context(), reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": subtasks})
}

func (h *SubtaskHandler) UpdateSubtask(c *gin.Context) {
	subtaskID := c.Param("subtask_id")

	var req models.UpdateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtask, err := h.service.Update(c.Request.Context(), subtaskID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": subtask})
}

func (h *SubtaskHandler) DeleteSubtask(c *gin.Context) {
	subtaskID := c.Param("subtask_id")

	if err := h.service.Delete(c.Request.Context(), subtaskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *SubtaskHandler) ToggleSubtask(c *gin.Context) {
	subtaskID := c.Param("subtask_id")

	subtask, err := h.service.ToggleComplete(c.Request.Context(), subtaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": subtask})
}
