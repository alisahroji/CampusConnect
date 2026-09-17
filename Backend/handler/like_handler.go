package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	service service.LikeService
}

func NewLikeHandler(service service.LikeService) *LikeHandler {
	return &LikeHandler{service: service}
}

// ToggleLike membalik status like untuk project yang sedang login (terproteksi)
// POST /api/projects/:id/like
func (h *LikeHandler) Toggle(c *gin.Context) {
	projectID := c.Param("id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	liked, likeCount, err := h.service.ToggleLike(projectID, userID.(string))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui like"})
		return
	}

	message := "Project berhasil di-like"
	if !liked {
		message = "Like project dibatalkan"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"liked":      liked,
		"like_count": likeCount,
	})
}

// GetLikeCount menampilkan total like sebuah project (publik, auth opsional).
// Bila request membawa token valid, response juga berisi current_user.liked
// agar frontend bisa mensinkronkan status like setelah refresh.
// GET /api/projects/:id/likes
func (h *LikeHandler) GetLikes(c *gin.Context) {
	projectID := c.Param("id")

	userID, exists := c.Get("userID")
	viewerID := ""
	if exists {
		viewerID, _ = userID.(string)
	}

	likeCount, liked, err := h.service.GetLikeInfo(projectID, viewerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil jumlah like"})
		return
	}

	resp := gin.H{
		"message":    "Berhasil",
		"like_count": likeCount,
	}

	// Tambahkan current_user hanya bila request terautentikasi
	// (tidak mengubah kontrak response untuk anonymous).
	if viewerID != "" {
		resp["current_user"] = gin.H{"liked": liked}
	}

	c.JSON(http.StatusOK, resp)
}
