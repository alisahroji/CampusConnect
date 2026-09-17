package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	service service.CommentService
}

func NewCommentHandler(service service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

// CommentInput struct untuk membuat komentar baru
type CommentInput struct {
	Content string `json:"content" binding:"required"`
}

// GetByProject menampilkan daftar komentar milik sebuah project (publik)
// GET /api/projects/:id/comments
func (h *CommentHandler) GetByProject(c *gin.Context) {
	projectID := c.Param("id")

	comments, err := h.service.GetComments(projectID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar komentar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil",
		"data":    comments,
	})
}

// Create menambahkan komentar baru pada sebuah project (terproteksi)
// POST /api/projects/:id/comments
func (h *CommentHandler) Create(c *gin.Context) {
	projectID := c.Param("id")

	var input CommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Isi komentar tidak boleh kosong"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	comment, err := h.service.AddComment(projectID, userID.(string), input.Content)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Komentar berhasil ditambahkan",
		"data":    comment,
	})
}

// Delete menghapus komentar milik user yang sedang login (terproteksi)
// DELETE /api/comments/:id
func (h *CommentHandler) Delete(c *gin.Context) {
	commentID := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteComment(commentID, userID.(string))
	if err != nil {
		if errors.Is(err, repository.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk menghapus komentar ini"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Komentar tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Komentar berhasil dihapus",
	})
}
