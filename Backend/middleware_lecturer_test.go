package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TEST LECTURER GUARD (Minggu 7 Day 2)
// AdminGuard diuji di middleware_test.go — pola sama (unit murni, tanpa DB).
// =============================================================================

func setupLecturerRouter(role string, banned bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		// Simulasi hasil RequireAuth: currentUser dimuat dari DB.
		c.Set("currentUser", User{ID: "u1", Role: role, Banned: banned})
		c.Next()
	})
	r.Use(LecturerGuard())
	r.POST("/api/materials", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestLecturerGuard_LecturerAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/materials", nil)
	setupLecturerRouter("Lecturer", false).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLecturerGuard_StudentForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/materials", nil)
	setupLecturerRouter("Student", false).ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLecturerGuard_AdminForbidden(t *testing.T) {
	// Policy Day 2: Admin tidak otomatis boleh upload materi.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/materials", nil)
	setupLecturerRouter("Admin", false).ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLecturerGuard_BannedLecturerForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/materials", nil)
	setupLecturerRouter("Lecturer", true).ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLecturerGuard_SetsUserRoleContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("currentUser", User{ID: "u1", Role: "Lecturer"})
		c.Next()
	})
	r.Use(LecturerGuard())
	var captured string
	r.POST("/api/materials", func(c *gin.Context) {
		captured = c.GetString("userRole")
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/materials", nil))
	assert.Equal(t, "Lecturer", captured)
}
