package api

import (
	"net/http"

	"reminder-service/internal/models"
	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	service *service.NoteService
}

func NewNoteHandler(svc *service.NoteService) *NoteHandler {
	return &NoteHandler{service: svc}
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
	reminderID := c.Param("id")
	userID := c.Query("user_id")

	var req models.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Create(c.Request.Context(), reminderID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": note})
}

func (h *NoteHandler) GetNotes(c *gin.Context) {
	reminderID := c.Param("id")
	notes, err := h.service.GetByReminder(c.Request.Context(), reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": notes})
}

func (h *NoteHandler) UpdateNote(c *gin.Context) {
	noteID := c.Param("note_id")
	var req models.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.Update(c.Request.Context(), noteID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": note})
}

func (h *NoteHandler) DeleteNote(c *gin.Context) {
	noteID := c.Param("note_id")
	if err := h.service.Delete(c.Request.Context(), noteID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
