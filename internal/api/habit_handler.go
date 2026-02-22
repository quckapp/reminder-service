package api

import (
	"net/http"
	"strconv"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type HabitHandler struct {
	svc *service.HabitService
}

func NewHabitHandler(svc *service.HabitService) *HabitHandler {
	return &HabitHandler{svc: svc}
}

func (h *HabitHandler) CreateHabit(c *gin.Context) {
	userID := c.Query("user_id")
	workspaceID := c.Query("workspace_id")

	var req models.CreateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	habit, err := h.svc.Create(c.Request.Context(), userID, workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": habit})
}

func (h *HabitHandler) GetHabit(c *gin.Context) {
	id := c.Param("id")

	habit, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": habit})
}

func (h *HabitHandler) ListHabits(c *gin.Context) {
	userID := c.Param("user_id")
	status := c.Query("status")

	habits, err := h.svc.ListByUser(c.Request.Context(), userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": habits})
}

func (h *HabitHandler) UpdateHabit(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	habit, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": habit})
}

func (h *HabitHandler) DeleteHabit(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HabitHandler) CompleteHabit(c *gin.Context) {
	habitID := c.Param("id")
	userID := c.Query("user_id")

	var req models.HabitCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body for simple completion
		req = models.HabitCompletionRequest{Count: 1}
	}

	completion, err := h.svc.Complete(c.Request.Context(), habitID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": completion})
}

func (h *HabitHandler) GetCompletions(c *gin.Context) {
	habitID := c.Param("id")
	limit := int64(50)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	completions, err := h.svc.GetCompletions(c.Request.Context(), habitID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": completions})
}

func (h *HabitHandler) GetStats(c *gin.Context) {
	habitID := c.Param("id")

	stats, err := h.svc.GetStats(c.Request.Context(), habitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *HabitHandler) GetSummary(c *gin.Context) {
	userID := c.Param("user_id")

	summary, err := h.svc.GetSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

func (h *HabitHandler) ResetStreak(c *gin.Context) {
	habitID := c.Param("id")

	if err := h.svc.ResetStreak(c.Request.Context(), habitID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
