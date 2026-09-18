package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NotificationHandler menangani endpoint notification dasar (Minggu 6).
// Storage-only: badge di frontend memakai polling, realtime menyusul Minggu 10.
type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// List: GET /api/notifications (RequireAuth)
// Mengembalikan daftar notification milik user yang login beserta
// unread_count untuk badge. Isolasi dijamin service/repository
// (query selalu dibatasi recipient = user dari token).
func (h *NotificationHandler) List(c *gin.Context) {
	userID := userIDFromContext(c)

	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}

	notifications, err := h.service.List(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil notifikasi"})
		return
	}

	unread, err := h.service.UnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung notifikasi belum dibaca"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         notifications,
		"unread_count": unread,
	})
}

// UnreadCount: GET /api/notifications/unread-count (RequireAuth)
// Endpoint ringan yang dipakai navbar badge untuk polling.
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	unread, err := h.service.UnreadCount(userIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung notifikasi belum dibaca"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": unread})
}

// MarkRead: POST /api/notifications/:id/read (RequireAuth)
// Hanya notification milik user yang login yang bisa ditandai (isolation).
// Menandai notifikasi yang sudah terbaca tetap sukses (idempoten).
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	err := h.service.MarkRead(c.Param("id"), userIDFromContext(c))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notifikasi tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menandai notifikasi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notifikasi ditandai terbaca"})
}

// MarkAllRead: POST /api/notifications/read-all (RequireAuth)
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	if err := h.service.MarkAllRead(userIDFromContext(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menandai semua notifikasi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Semua notifikasi ditandai terbaca"})
}
