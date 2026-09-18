package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BookmarkHandler menangani endpoint bookmark Project & Post (Minggu 6).
// Mapping status konsisten dengan handler lain:
//   - repository.ErrNotFound -> 404
//   - selain itu             -> 500 (tanpa membocorkan detail internal)
type BookmarkHandler struct {
	service service.BookmarkService
}

func NewBookmarkHandler(service service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{service: service}
}

// userIDFromContext mengambil userID hasil middleware RequireAuth/OptionalAuth.
// Mengembalikan "" bila konteks tidak membawa user (anonymous).
func userIDFromContext(c *gin.Context) string {
	val, exists := c.Get("userID")
	if !exists {
		return ""
	}
	id, _ := val.(string)
	return id
}

// ToggleProjectBookmark: POST /api/projects/:id/bookmark (RequireAuth)
func (h *BookmarkHandler) ToggleProjectBookmark(c *gin.Context) {
	projectID := c.Param("id")

	bookmarked, err := h.service.ToggleProjectBookmark(projectID, userIDFromContext(c))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui bookmark"})
		return
	}

	message := "Project berhasil disimpan"
	if !bookmarked {
		message = "Bookmark project dihapus"
	}
	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"bookmarked": bookmarked,
	})
}

// GetProjectBookmarkStatus: GET /api/projects/:id/bookmark (OptionalAuth)
func (h *BookmarkHandler) GetProjectBookmarkStatus(c *gin.Context) {
	projectID := c.Param("id")

	bookmarked, err := h.service.GetProjectBookmarkStatus(projectID, userIDFromContext(c))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil status bookmark"})
		return
	}

	resp := gin.H{"bookmarked": bookmarked}
	if userIDFromContext(c) != "" {
		resp["current_user"] = gin.H{"bookmarked": bookmarked}
	}
	c.JSON(http.StatusOK, resp)
}

// ListProjectBookmarks: GET /api/bookmarks/projects (RequireAuth)
func (h *BookmarkHandler) ListProjectBookmarks(c *gin.Context) {
	projects, err := h.service.ListProjectBookmarks(userIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar bookmark project"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// TogglePostBookmark: POST /api/posts/:id/bookmark (RequireAuth)
func (h *BookmarkHandler) TogglePostBookmark(c *gin.Context) {
	postID := c.Param("id")

	bookmarked, err := h.service.TogglePostBookmark(postID, userIDFromContext(c))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui bookmark"})
		return
	}

	message := "Post berhasil disimpan"
	if !bookmarked {
		message = "Bookmark post dihapus"
	}
	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"bookmarked": bookmarked,
	})
}

// GetPostBookmarkStatus: GET /api/posts/:id/bookmark (OptionalAuth)
func (h *BookmarkHandler) GetPostBookmarkStatus(c *gin.Context) {
	postID := c.Param("id")

	bookmarked, err := h.service.GetPostBookmarkStatus(postID, userIDFromContext(c))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil status bookmark"})
		return
	}

	resp := gin.H{"bookmarked": bookmarked}
	if userIDFromContext(c) != "" {
		resp["current_user"] = gin.H{"bookmarked": bookmarked}
	}
	c.JSON(http.StatusOK, resp)
}

// ListPostBookmarks: GET /api/bookmarks/posts (RequireAuth)
func (h *BookmarkHandler) ListPostBookmarks(c *gin.Context) {
	posts, err := h.service.ListPostBookmarks(userIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar bookmark post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": posts})
}
