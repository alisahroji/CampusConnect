package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// adminUserView adalah DTO daftar user untuk admin (Minggu 6 Hari 4).
// Field eksplisit: TIDAK menyertakan field sensitif apa pun. Struct User
// memang tidak menyimpan password/OTP/token, tapi DTO eksplisit menjamin
// field baru yang sensitif tidak ikut ter-expose otomatis.
type adminUserView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Banned    bool      `json:"banned"`
	CreatedAt time.Time `json:"created_at"`
}

func toAdminUserViews(users []repository.User) []adminUserView {
	views := make([]adminUserView, 0, len(users))
	for _, u := range users {
		views = append(views, adminUserView{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      u.Role,
			Banned:    u.Banned,
			CreatedAt: u.CreatedAt,
		})
	}
	return views
}

// AdminHandler menangani endpoint admin (Minggu 6 Hari 4).
// Semua rute dipagari RequireAuth + AdminGuard di main.go.
type AdminHandler struct {
	service service.AdminService
}

func NewAdminHandler(service service.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

// ListUsers: GET /api/admin/users (RequireAuth + AdminGuard)
func (h *AdminHandler) ListUsers(c *gin.Context) {
	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}

	users, err := h.service.ListUsers(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toAdminUserViews(users)})
}

// SetBanned: POST /api/admin/users/:id/ban dan .../unban (RequireAuth + AdminGuard)
func (h *AdminHandler) SetBanned(banned bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorID := userIDFromContext(c)
		targetID := c.Param("id")

		if err := h.service.SetBanned(actorID, targetID, banned); err != nil {
			if errors.Is(err, service.ErrSelfBan) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status user"})
			return
		}

		message := "User berhasil diblokir"
		if !banned {
			message = "Blokir user berhasil dilepas"
		}
		c.JSON(http.StatusOK, gin.H{"message": message, "banned": banned})
	}
}

// SetRole: POST /api/admin/users/:id/role (RequireAuth + AdminGuard)
func (h *AdminHandler) SetRole(c *gin.Context) {
	actorID := userIDFromContext(c)
	targetID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Field role wajib diisi"})
		return
	}

	updated, err := h.service.SetRole(actorID, targetID, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRole) || errors.Is(err, service.ErrLastAdmin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengubah role user"})
		return
	}

	// Response hanya field aman (DTO admin, tanpa field sensitif).
	c.JSON(http.StatusOK, gin.H{
		"message": "Role berhasil diperbarui",
		"data":    toAdminUserViews([]repository.User{*updated})[0],
	})
}
