package api

import (
	"net/http"
	"strconv"

	"reminder-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Extended2Handler struct {
	svc *service.Extended2Service
}

func NewExtended2Handler(svc *service.Extended2Service) *Extended2Handler {
	return &Extended2Handler{svc: svc}
}

func ext2Limit(c *gin.Context) int {
	l, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if l <= 0 || l > 200 { l = 50 }
	return l
}

func ext2Offset(c *gin.Context) int {
	o, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if o < 0 { o = 0 }
	return o
}

func ext2UserID(c *gin.Context) string {
	uid := c.GetHeader("X-User-ID")
	if uid == "" { uid = c.Query("user_id") }
	return uid
}

// ── Attachments ──

func (h *Extended2Handler) AddAttachment(c *gin.Context) {
	var req struct {
		FileName string `json:"file_name" binding:"required"`
		FileURL  string `json:"file_url" binding:"required"`
		MimeType string `json:"mime_type"`
		Size     int64  `json:"size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	att := &service.ReminderAttachment{
		ReminderID: c.Param("id"),
		FileName:   req.FileName,
		FileURL:    req.FileURL,
		MimeType:   req.MimeType,
		Size:       req.Size,
		UploadedBy: ext2UserID(c),
	}
	if err := h.svc.AddAttachment(c.Request.Context(), att); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": att})
}

func (h *Extended2Handler) ListAttachments(c *gin.Context) {
	results, err := h.svc.ListAttachments(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) DeleteAttachment(c *gin.Context) {
	if err := h.svc.DeleteAttachment(c.Request.Context(), c.Param("attachmentId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Comments ──

func (h *Extended2Handler) AddComment(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment := &service.ReminderComment{
		ReminderID: c.Param("id"),
		UserID:     ext2UserID(c),
		Content:    req.Content,
	}
	if err := h.svc.AddComment(c.Request.Context(), comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": comment})
}

func (h *Extended2Handler) ListComments(c *gin.Context) {
	results, err := h.svc.ListComments(c.Request.Context(), c.Param("id"), ext2Limit(c), ext2Offset(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) UpdateComment(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateComment(c.Request.Context(), c.Param("commentId"), req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) DeleteComment(c *gin.Context) {
	if err := h.svc.DeleteComment(c.Request.Context(), c.Param("commentId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Reactions ──

func (h *Extended2Handler) AddReaction(c *gin.Context) {
	var req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r := &service.ReminderReaction{
		ReminderID: c.Param("id"),
		UserID:     ext2UserID(c),
		Emoji:      req.Emoji,
	}
	if err := h.svc.AddReaction(c.Request.Context(), r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": r})
}

func (h *Extended2Handler) RemoveReaction(c *gin.Context) {
	if err := h.svc.RemoveReaction(c.Request.Context(), c.Param("id"), ext2UserID(c), c.Query("emoji")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) ListReactions(c *gin.Context) {
	results, err := h.svc.ListReactions(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ── Watchers ──

func (h *Extended2Handler) AddWatcher(c *gin.Context) {
	w := &service.ReminderWatcher{
		ReminderID: c.Param("id"),
		UserID:     ext2UserID(c),
	}
	if err := h.svc.AddWatcher(c.Request.Context(), w); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": w})
}

func (h *Extended2Handler) RemoveWatcher(c *gin.Context) {
	if err := h.svc.RemoveWatcher(c.Request.Context(), c.Param("id"), ext2UserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) ListWatchers(c *gin.Context) {
	results, err := h.svc.ListWatchers(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ── Labels ──

func (h *Extended2Handler) AddLabel(c *gin.Context) {
	var req struct {
		Label string `json:"label" binding:"required"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l := &service.ReminderLabel{
		ReminderID: c.Param("id"),
		Label:      req.Label,
		Color:      req.Color,
	}
	if err := h.svc.AddLabel(c.Request.Context(), l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": l})
}

func (h *Extended2Handler) RemoveLabel(c *gin.Context) {
	if err := h.svc.RemoveLabel(c.Request.Context(), c.Param("id"), c.Param("label")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) ListLabels(c *gin.Context) {
	results, err := h.svc.ListLabels(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) SearchByLabel(c *gin.Context) {
	results, err := h.svc.SearchByLabel(c.Request.Context(), c.Param("label"), ext2Limit(c), ext2Offset(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ── Favorites ──

func (h *Extended2Handler) AddFavorite(c *gin.Context) {
	f := &service.ReminderFavorite{
		ReminderID: c.Param("id"),
		UserID:     ext2UserID(c),
	}
	if err := h.svc.AddFavorite(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": f})
}

func (h *Extended2Handler) RemoveFavorite(c *gin.Context) {
	if err := h.svc.RemoveFavorite(c.Request.Context(), c.Param("id"), ext2UserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) ListFavorites(c *gin.Context) {
	results, err := h.svc.ListFavorites(c.Request.Context(), c.Param("user_id"), ext2Limit(c), ext2Offset(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) IsFavorited(c *gin.Context) {
	fav, err := h.svc.IsFavorited(c.Request.Context(), c.Param("id"), ext2UserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "favorited": fav})
}

// ── Dependencies ──

func (h *Extended2Handler) AddDependency(c *gin.Context) {
	var req struct {
		DependsOnID string `json:"depends_on_id" binding:"required"`
		Type        string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d := &service.ReminderDependency{
		ReminderID:  c.Param("id"),
		DependsOnID: req.DependsOnID,
		Type:        req.Type,
	}
	if err := h.svc.AddDependency(c.Request.Context(), d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": d})
}

func (h *Extended2Handler) RemoveDependency(c *gin.Context) {
	if err := h.svc.RemoveDependency(c.Request.Context(), c.Param("depId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) ListDependencies(c *gin.Context) {
	results, err := h.svc.ListDependencies(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ── Locations ──

func (h *Extended2Handler) SetLocation(c *gin.Context) {
	var req struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Radius    float64 `json:"radius"`
		TriggerOn string  `json:"trigger_on"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	loc := &service.ReminderLocation{
		ReminderID: c.Param("id"),
		Name:       req.Name,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Radius:     req.Radius,
		TriggerOn:  req.TriggerOn,
	}
	if err := h.svc.SetLocation(c.Request.Context(), loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": loc})
}

func (h *Extended2Handler) GetLocation(c *gin.Context) {
	loc, err := h.svc.GetLocation(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": loc})
}

func (h *Extended2Handler) RemoveLocation(c *gin.Context) {
	if err := h.svc.RemoveLocation(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) ListNearby(c *gin.Context) {
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lon, _ := strconv.ParseFloat(c.Query("lon"), 64)
	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "1000"), 64)
	results, err := h.svc.ListNearby(c.Request.Context(), lat, lon, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ── Snooze History ──

func (h *Extended2Handler) ListSnoozeHistory(c *gin.Context) {
	results, err := h.svc.ListSnoozeHistory(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// ── Quick Actions ──

func (h *Extended2Handler) CreateQuickAction(c *gin.Context) {
	var req struct {
		Name   string                 `json:"name" binding:"required"`
		Action string                 `json:"action" binding:"required"`
		Config map[string]interface{} `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	qa := &service.ReminderQuickAction{
		UserID:      ext2UserID(c),
		WorkspaceID: c.Query("workspace_id"),
		Name:        req.Name,
		Action:      req.Action,
		Config:      req.Config,
	}
	if err := h.svc.CreateQuickAction(c.Request.Context(), qa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": qa})
}

func (h *Extended2Handler) ListQuickActions(c *gin.Context) {
	results, err := h.svc.ListQuickActions(c.Request.Context(), ext2UserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) DeleteQuickAction(c *gin.Context) {
	if err := h.svc.DeleteQuickAction(c.Request.Context(), c.Param("actionId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Stats ──

func (h *Extended2Handler) GetCompletionRate(c *gin.Context) {
	result, err := h.svc.GetCompletionRate(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *Extended2Handler) GetStreakInfo(c *gin.Context) {
	result, err := h.svc.GetStreakInfo(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
