package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostInteractionHandler struct {
	likeService    service.PostLikeService
	commentService service.PostCommentService
}

func NewPostInteractionHandler(likeService service.PostLikeService, commentService service.PostCommentService) *PostInteractionHandler {
	return &PostInteractionHandler{likeService: likeService, commentService: commentService}
}

// PostCommentInput struct untuk membuat komentar post baru
type PostCommentInput struct {
	Content string `json:"content" binding:"required"`
}

// ToggleLike membalik status like untuk post yang sedang login (terproteksi)
// POST /api/posts/:id/like
func (h *PostInteractionHandler) ToggleLike(c *gin.Context) {
	postID := c.Param("id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	liked, likeCount, err := h.likeService.TogglePostLike(postID, userID.(string))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui like"})
		return
	}

	message := "Post berhasil di-like"
	if !liked {
		message = "Like post dibatalkan"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"liked":      liked,
		"like_count": likeCount,
	})
}

// GetLikes menampilkan total like sebuah post (publik)
// GET /api/posts/:id/likes
func (h *PostInteractionHandler) GetLikes(c *gin.Context) {
	postID := c.Param("id")

	likeCount, err := h.likeService.GetPostLikeCount(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil jumlah like"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Berhasil",
		"like_count": likeCount,
	})
}

// GetComments menampilkan daftar komentar milik sebuah post (publik)
// GET /api/posts/:id/comments
func (h *PostInteractionHandler) GetComments(c *gin.Context) {
	postID := c.Param("id")

	comments, err := h.commentService.GetPostComments(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
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

// CreateComment menambahkan komentar baru pada sebuah post (terproteksi)
// POST /api/posts/:id/comments
func (h *PostInteractionHandler) CreateComment(c *gin.Context) {
	postID := c.Param("id")

	var input PostCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Isi komentar tidak boleh kosong"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	comment, err := h.commentService.AddPostComment(postID, userID.(string), input.Content)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post tidak ditemukan"})
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

// DeleteComment menghapus komentar post milik user yang sedang login (terproteksi)
// DELETE /api/post-comments/:id
func (h *PostInteractionHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.commentService.DeletePostComment(commentID, userID.(string))
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
