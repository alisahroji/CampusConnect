package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	service service.PostService
}

func NewPostHandler(service service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

// PostInput struct untuk membuat post baru
type PostInput struct {
	Content  string `json:"content" binding:"required"`
	ImageURL string `json:"image_url"`
}

// Create menambahkan post baru pada feed (terproteksi)
// POST /api/posts
func (h *PostHandler) Create(c *gin.Context) {
	var input PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konten post wajib diisi"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	post := &repository.Post{
		UserID:   userID.(string),
		Content:  input.Content,
		ImageURL: input.ImageURL,
	}

	if err := h.service.CreatePost(post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post berhasil dibuat",
		"data":    post,
	})
}

// getPostViewerID mengambil userID dari context bila request membawa token
// valid (auth opsional). Bila anonymous, mengembalikan string kosong.
func getPostViewerID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	id, ok := userID.(string)
	if !ok {
		return ""
	}
	return id
}

// GetAll menampilkan daftar post untuk feed explore (publik).
// GET /api/posts?cursor=<base64>&limit=10
// Pagination cursor-based (keyset) untuk mendukung infinite scrolling.
// Bila request membawa token valid (OptionalAuth), setiap post juga memuat
// liked_by_me agar ikon like frontend tetap sinkron setelah refresh.
func (h *PostHandler) GetAll(c *gin.Context) {
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, nextCursor, err := h.service.GetAllPosts(cursor, limit, getPostViewerID(c))
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCursor) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cursor tidak valid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Berhasil",
		"data":        posts,
		"next_cursor": nextCursor, // "" berarti sudah halaman terakhir
	})
}

// GetByID menampilkan detail satu post (publik, auth opsional)
// GET /api/posts/:id
func (h *PostHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	post, err := h.service.GetPostByID(id, getPostViewerID(c))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil",
		"data":    post,
	})
}

// Update memperbarui post milik user yang sedang login (terproteksi).
// PUT /api/posts/:id — konsisten dengan PUT /api/projects/:id.
func (h *PostHandler) Update(c *gin.Context) {
	postID := c.Param("id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var input PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konten post wajib diisi"})
		return
	}

	updates := &repository.Post{
		Content:  input.Content,
		ImageURL: input.ImageURL,
	}

	updated, err := h.service.UpdatePost(postID, userID.(string), updates)
	if err != nil {
		if errors.Is(err, repository.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk mengubah post ini"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		// Selain itu error validasi bisnis (kosong / melebihi 2000 karakter)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post berhasil diperbarui",
		"data":    updated,
	})
}

// Delete menghapus post milik user yang sedang login (terproteksi)
// DELETE /api/posts/:id
func (h *PostHandler) Delete(c *gin.Context) {
	postID := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeletePost(postID, userID.(string))
	if err != nil {
		// Map sentinel errors ke status HTTP yang tepat
		if errors.Is(err, repository.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk menghapus post ini"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post berhasil dihapus",
	})
}
