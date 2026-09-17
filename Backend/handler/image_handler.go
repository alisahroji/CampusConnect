package handler

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

// ImageHandler menangani upload gambar galeri project ke Cloudinary.
// Handler ini HANYA mengunggah file dan mengembalikan Secure URL-nya;
// penyimpanan URL ke galeri project dilakukan lewat POST/PUT /api/projects
// (field gallery) agar ownership project tetap tervalidasi di service.
type ImageHandler struct{}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

// UploadProjectImage mengunggah satu gambar (multipart field "image")
// ke folder Cloudinary campusconnect/projects dan mengembalikan URL publiknya.
// POST /api/images/project (terproteksi RequireAuth)
func (h *ImageHandler) UploadProjectImage(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File gambar tidak ditemukan (field 'image')"})
		return
	}
	defer file.Close()

	// Validasi dasar: hanya menerima content-type gambar
	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg":               true,
		"image/png":                true,
		"image/gif":                true,
		"image/webp":               true,
		"application/octet-stream": true, // beberapa browser kirim generik untuk file gambar
	}
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file harus berupa gambar (JPG/PNG/GIF/WebP)"})
		return
	}

	// Batas ukuran 5 MB per gambar galeri
	const maxImageSize = 5 * 1024 * 1024
	if fileHeader.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran gambar maksimal 5 MB"})
		return
	}

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghubungkan ke server gambar"})
		return
	}

	// Timeout 30 detik agar request tidak menggantung selamanya
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "campusconnect/projects",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengunggah gambar ke cloud"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Gambar berhasil diunggah",
		"image_url": uploadResult.SecureURL,
	})
}
