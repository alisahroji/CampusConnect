package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"context"
	"errors"
	"net/http"
	"os"

	"fmt"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService   service.UserService
	followService service.FollowService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

// SetFollowService menyuntikkan dependensi FollowService untuk endpoint
// profil publik (GetPublicProfile). Dipisah dari konstruktor agar wiring
// di main.go tidak perlu diubah urutannya (followService dibuat setelah
// userHandler).
func (h *UserHandler) SetFollowService(followService service.FollowService) {
	h.followService = followService
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	// Ambil data user yang disimpan oleh middleware RequireAuth dari context
	userContext, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: User tidak ditemukan di sistem"})
		return
	}

	// Karena userContext bertipe umum (interface{}), kita kirim langsung tanpa cast manual
	// atau gunakan tipe data map/struct yang sesuai.
	// Cara paling aman dan bersih di handler: langsung kembalikan userContext dari middleware.

	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil dimuat via Clean Architecture!",
		"user":    userContext,
	})
}

// GetPublicProfile menampilkan profil publik seorang user berdasarkan ID
// (dipakai halaman UserProfile untuk Follow UI). Bila viewer membawa token
// valid, response juga memuat current_user.following agar tombol Follow
// tetap sinkron dengan server setelah refresh.
// GET /api/users/:id
func (h *UserHandler) GetPublicProfile(c *gin.Context) {
	userID := c.Param("id")

	viewerID := ""
	if viewer, exists := c.Get("userID"); exists {
		viewerID, _ = viewer.(string)
	}

	user, following, err := h.followService.GetUserWithFollowStatus(userID, viewerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat profil user"})
		return
	}

	// Email dianggap data privat: tidak pernah bocor lewat endpoint publik.
	publicUser := *user
	publicUser.Email = ""

	resp := gin.H{
		"message": "Berhasil",
		"user":    publicUser,
	}
	if viewerID != "" && viewerID != userID {
		resp["current_user"] = gin.H{"following": following}
	}

	c.JSON(http.StatusOK, resp)
}

// Struct untuk menampung data JSON yang dikirim dari Frontend/Thunder Client
type UpdateProfileInput struct {
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	Skills      string `json:"skills"`
	GithubURL   string `json:"github_url"`
	LinkedinURL string `json:"linkedin_url"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 1. Ambil ID User dari satpam (middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	// 2. Ambil data asli user dari database menggunakan Service
	user, err := h.userService.GetUserProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan di database"})
		return
	}

	// 3. Tangkap data JSON yang dikirimkan via Body
	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid"})
		return
	}

	// 4. Timpa data lama dengan data baru
	// Kita biarkan Email dan Role tetap (tidak boleh diubah sembarangan)
	user.Name = input.Name
	user.Bio = input.Bio
	user.Skills = input.Skills
	user.GithubURL = input.GithubURL
	user.LinkedinURL = input.LinkedinURL

	// 5. Simpan perubahan ke database via Service
	if err := h.userService.UpdateProfile(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan perubahan profil"})
		return
	}

	// 6. Beri respon sukses
	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil diperbarui!",
		"user":    user,
	})
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	fmt.Println("[DEBUG] 1. Masuk ke fungsi UploadAvatar")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	fmt.Println("[DEBUG] 2. Membaca file dari form...")
	file, fileHeader, err := c.Request.FormFile("avatar")
	if err != nil {
		fmt.Println("[ERROR] Gagal membaca file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca file gambar."})
		return
	}
	defer file.Close()
	fmt.Println("[DEBUG] File terbaca:", fileHeader.Filename, "| Ukuran:", fileHeader.Size, "bytes")

	fmt.Println("[DEBUG] 3. Menghubungkan ke Cloudinary...")
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		fmt.Println("[ERROR] Kredensial Cloudinary salah:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungkan ke server gambar"})
		return
	}

	fmt.Println("[DEBUG] 4. Mulai proses upload ke server Cloudinary (ini butuh internet)...")
	// Tambahkan timeout 15 detik agar tidak nge-hang selamanya
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "campusconnect/avatars",
	})
	if err != nil {
		fmt.Println("[ERROR] Cloudinary menolak upload:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengunggah gambar ke cloud"})
		return
	}

	fmt.Println("[DEBUG] 5. Upload sukses! URL didapat:", uploadResult.SecureURL)

	user, err := h.userService.GetUserProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	user.PictureURL = uploadResult.SecureURL

	if err := h.userService.UpdateProfile(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan URL gambar ke database"})
		return
	}

	fmt.Println("[DEBUG] 6. Selesai! Mengirim respon 200 OK ke Thunder Client")
	c.JSON(http.StatusOK, gin.H{
		"message":     "Foto profil berhasil diunggah!",
		"picture_url": user.PictureURL,
	})
}
