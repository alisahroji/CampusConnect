package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	service service.FollowService
}

func NewFollowHandler(service service.FollowService) *FollowHandler {
	return &FollowHandler{service: service}
}

// Toggle membalik status follow terhadap user lain (terproteksi)
// POST /api/users/:id/follow
func (h *FollowHandler) Toggle(c *gin.Context) {
	targetUserID := c.Param("id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	following, err := h.service.ToggleFollow(userID.(string), targetUserID)
	if err != nil {
		// Target user tidak ada → 404
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User yang akan di-follow tidak ditemukan"})
			return
		}
		// Self-follow dan format ID invalid ditolak di service (400)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := "Berhasil follow user"
	if !following {
		message = "Berhasil unfollow user"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   message,
		"following": following,
	})
}

// GetFeed menampilkan post dari user yang di-follow + post sendiri (terproteksi)
// GET /api/feed
type FeedHandler struct {
	service service.FeedService
}

func NewFeedHandler(service service.FeedService) *FeedHandler {
	return &FeedHandler{service: service}
}

func (h *FeedHandler) GetFeed(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, nextCursor, err := h.service.GetFeed(userID.(string), cursor, limit)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCursor) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cursor tidak valid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat feed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Berhasil",
		"data":        posts,
		"next_cursor": nextCursor, // "" berarti sudah halaman terakhir
	})
}
