package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

// MaterialHandler menangani HTTP untuk CRUD Material (Minggu 7 Day 2).
// Business rule ada di MaterialService; handler hanya parse request + response.
type MaterialHandler struct {
	service service.MaterialService
}

func NewMaterialHandler(service service.MaterialService) *MaterialHandler {
	return &MaterialHandler{service: service}
}

// NewCloudinaryUploader dikonsumsi oleh main.go saat wiring MaterialService.
func NewCloudinaryUploader() service.FileUploader {
	return cloudinaryUploader{}
}

// cloudinaryUploader mengimplementasikan service.FileUploader dengan Cloudinary.
// CATATAN RIIL (terverifikasi empiris Day 3): SDK v2.16 selalu POST ke endpoint
// `auto` (BuildPath(api.Auto, ...)) sehingga field ResourceType TIDAK BERDAMPAK —
// tipe akhir ditentukan deteksi server Cloudinary: txt/docx/xlsx -> /raw/upload/,
// pdf -> /image/upload/. Delivery PDF juga dibatasi pengaturan keamanan akun
// (X-Cld-Error: "deny or ACL failure", HTTP 401) — mengaktifkan "PDF and ZIP
// files delivery" di dashboard Cloudinary adalah fix yang benar (bukan kode).
// Raw (txt/dokumen office) terbukti 200 publik.
type cloudinaryUploader struct{}

func (cloudinaryUploader) Upload(filename string, size int64, content io.Reader) (string, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", err
	}

	// Timeout 60 detik: dokumen lebih besar dari gambar, jangan biarkan hang.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	uploadResult, err := cld.Upload.Upload(ctx, content, uploader.UploadParams{
		Folder: "campusconnect/materials",
		// ResourceType tidak dikirim: endpoint SDK adalah `auto`, deteksi tipe
		// dilakukan server (lihat CATATAN RIIL di atas).
	})
	if err != nil {
		return "", err
	}
	return uploadResult.SecureURL, nil
}

// ---------------------------------------------------------------------------
// DTO: tidak mengekspos struct repository.User penuh (pola DTO admin Day 4 W6)
// ---------------------------------------------------------------------------

type materialUploaderView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PictureURL string `json:"picture_url"`
}

type materialView struct {
	ID               string               `json:"id"`
	Title            string               `json:"title"`
	Category         string               `json:"category"`
	FileURL          string               `json:"file_url"`
	OriginalFilename string               `json:"original_filename"`
	FileSize         int64                `json:"file_size"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	Uploader         materialUploaderView `json:"uploader"`
}

func toMaterialView(m *repository.Material) materialView {
	view := materialView{
		ID:               m.ID,
		Title:            m.Title,
		Category:         m.Category,
		FileURL:          m.FileURL,
		OriginalFilename: m.OriginalFilename,
		FileSize:         m.FileSize,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.Uploader != nil {
		view.Uploader = materialUploaderView{
			ID:         m.Uploader.ID,
			Name:       m.Uploader.Name,
			PictureURL: m.Uploader.PictureURL,
		}
	}
	return view
}

func toMaterialViews(list []repository.Material) []materialView {
	views := make([]materialView, 0, len(list))
	for i := range list {
		views = append(views, toMaterialView(&list[i]))
	}
	return views
}

// ---------------------------------------------------------------------------
// Helper response/error
// ---------------------------------------------------------------------------

// requireUserID mengembalikan userID dari context RequireAuth, atau menulis
// response 401 dan mengembalikan false.
func requireUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return "", false
	}
	id, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return "", false
	}
	return id, true
}

// mapMaterialError menerjemahkan error service ke response HTTP.
// Semua error non-sentinel yang sampai di sini sudah dibungkus service dengan
// pesan aman (tanpa detail SQL/provider), mengikuti pola mapProjectError.
func mapMaterialError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Materi tidak ditemukan"})
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Create mengunggah materi baru — Lecturer only (LecturerGuard di route).
// POST /api/materials (multipart/form-data: title, category, file)
func (h *MaterialHandler) Create(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	// Defense in depth: role dibaca dari context yang di-set LecturerGuard
	// (sumber kebenaran DB), BUKAN dari payload client. Bila route lupa dipasang
	// guard, role kosong -> service juga menolak.
	if role := c.GetString("userRole"); role != "Lecturer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Endpoint ini khusus dosen (Lecturer)"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	category := strings.TrimSpace(c.PostForm("category"))

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File materi wajib dipilih (field 'file')"})
		return
	}
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca file"})
		return
	}
	defer src.Close()

	material, err := h.service.CreateMaterial(userID, "Lecturer", title, category, fileHeader.Filename, fileHeader.Size, src)
	if err != nil {
		mapMaterialError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Materi berhasil diunggah",
		"data":    toMaterialView(material),
	})
}

// GetAll menampilkan daftar materi (authenticated, semua role).
// GET /api/materials?category=&uploader_id=&limit=
func (h *MaterialHandler) GetAll(c *gin.Context) {
	if _, ok := requireUserID(c); !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.service.ListMaterials(c.Query("category"), c.Query("uploader_id"), limit)
	if err != nil {
		mapMaterialError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toMaterialViews(list)})
}

// GetByID menampilkan detail satu materi (authenticated, semua role).
// GET /api/materials/:id
func (h *MaterialHandler) GetByID(c *gin.Context) {
	if _, ok := requireUserID(c); !ok {
		return
	}
	material, err := h.service.GetMaterial(c.Param("id"))
	if err != nil {
		mapMaterialError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toMaterialView(material)})
}

// MaterialUpdateInput payload update metadata materi.
type MaterialUpdateInput struct {
	Title    string `json:"title" binding:"required"`
	Category string `json:"category" binding:"required"`
}

// Update mengubah metadata materi (judul/kategori) — hanya uploader.
// PUT /api/materials/:id
// Penggantian file sengaja DITUNDA (dokumentasi laporan Day 2): butuh
// integrasi Cloudinary Destroy yang belum ada di architecture existing.
func (h *MaterialHandler) Update(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var input MaterialUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}
	material, err := h.service.UpdateMaterial(c.Param("id"), userID, input.Title, input.Category)
	if err != nil {
		mapMaterialError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Materi berhasil diperbarui",
		"data":    toMaterialView(material),
	})
}

// Delete menghapus materi — hanya uploader.
// DELETE /api/materials/:id
func (h *MaterialHandler) Delete(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteMaterial(c.Param("id"), userID); err != nil {
		mapMaterialError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Materi berhasil dihapus"})
}
