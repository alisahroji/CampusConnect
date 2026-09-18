package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TEST ADMIN GUARD (Minggu 6 Hari 4)
// AdminGuard membaca currentUser hasil RequireAuth dari context — test di sini
// murni unit (tanpa DB): context di-set langsung sesuai kontrak RequireAuth.
// =============================================================================

func performGuardTest(user *User, setAuthed bool) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		// Simulasi hasil RequireAuth sebelum AdminGuard.
		if setAuthed && user != nil {
			c.Set("currentUser", *user)
			c.Set("userID", user.ID)
		}
	}, AdminGuard(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	router.ServeHTTP(w, req)
	return w
}

func TestAdminGuard_AdminAllowed(t *testing.T) {
	admin := &User{ID: "11111111-1111-1111-1111-111111111111", Role: "Admin"}
	w := performGuardTest(admin, true)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestAdminGuard_StudentForbidden(t *testing.T) {
	student := &User{ID: "22222222-2222-2222-2222-222222222222", Role: "Student"}
	w := performGuardTest(student, true)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "khusus admin")
}

func TestAdminGuard_LecturerForbidden(t *testing.T) {
	lecturer := &User{ID: "33333333-3333-3333-3333-333333333333", Role: "Lecturer"}
	w := performGuardTest(lecturer, true)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGuard_NoUserInContext_Unauthorized(t *testing.T) {
	// Pertahanan: jika RequireAuth tidak jalan, guard harus menolak (401).
	w := performGuardTest(nil, false)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminGuard_BannedAdminForbidden(t *testing.T) {
	// Banned admin tetap ditolak di guard (role benar tapi role hanya berlaku
	// untuk non-banned user — dibekukan juga oleh RequireAuth sebelum guard).
	bannedAdmin := &User{ID: "11111111-1111-1111-1111-111111111111", Role: "Admin", Banned: true}
	w := performGuardTest(bannedAdmin, true)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
