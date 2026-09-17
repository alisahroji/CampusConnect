package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service service.ProjectService
}

func NewProjectHandler(service service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// ProjectInput adalah struktur payload JSON untuk create/update project.
type ProjectInput struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	TechStack   string   `json:"tech_stack"`
	RepoURL     string   `json:"repo_url"`
	DemoURL     string   `json:"demo_url"`
	ImageURL    string   `json:"image_url"`
	Status      string   `json:"status"`
	Gallery     []string `json:"gallery"`
}

// getOptionalUserID mengambil userID dari context bila request membawa token
// valid (auth opsional). Bila anonymous, mengembalikan string kosong.
func getOptionalUserID(c *gin.Context) string {
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

// mapProjectError menerjemahkan sentinel error repository ke response HTTP
// dengan status code yang tepat. forbiddenMsg = pesan khusus untuk 403.
func mapProjectError(c *gin.Context, err error, forbiddenMsg string) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": forbiddenMsg})
		return
	}
	// Selain itu dianggap error validasi bisnis (mis. judul kosong, status tidak valid)
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// buildGallery mengubah daftar URL mentah dari payload menjadi row galeri siap simpan:
// trim + buang entri kosong. Urutan payload menentukan display_order.
func buildGallery(urls []string) []repository.ProjectGallery {
	gallery := make([]repository.ProjectGallery, 0, len(urls))
	for _, u := range urls {
		trimmed := strings.TrimSpace(u)
		if trimmed == "" {
			continue
		}
		gallery = append(gallery, repository.ProjectGallery{ImageURL: trimmed})
	}
	return gallery
}

// Create membuat project baru (terproteksi).
// POST /api/projects
func (h *ProjectHandler) Create(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var input ProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}

	project := &repository.Project{
		UserID:      userID.(string),
		Title:       input.Title,
		Description: input.Description,
		TechStack:   input.TechStack,
		RepoURL:     input.RepoURL,
		DemoURL:     input.DemoURL,
		ImageURL:    input.ImageURL,
		Status:      input.Status,
	}

	if err := h.service.CreateProject(project, buildGallery(input.Gallery)); err != nil {
		mapProjectError(c, err, "Anda tidak memiliki izin untuk membuat project ini")
		return
	}

	// Muat ulang project agar galeri ikut dalam response
	created, err := h.service.GetProjectByID(project.ID, userID.(string))
	if err != nil {
		created = project
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project berhasil dibuat",
		"data":    created,
	})
}

// GetAll menampilkan daftar project (publik).
// Mendukung filter: ?status=draft|published, ?tech_stack=React (filter nyata di
// database), dan ?mine=true untuk mengambil seluruh project milik user yang
// sedang login termasuk draft (butuh token).
func (h *ProjectHandler) GetAll(c *gin.Context) {
	viewerID := getOptionalUserID(c)

	// Mode "projects saya": seluruh project milik user yang sedang login
	if c.Query("mine") == "true" {
		if viewerID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
			return
		}
		projects, err := h.service.GetMyProjects(viewerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar project"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": projects})
		return
	}

	filter := repository.ProjectFilter{
		Status:    c.Query("status"),
		TechStack: strings.TrimSpace(c.Query("tech_stack")),
	}

	// Validasi parameter status: nilai tidak dikenal ditolak jelas (400),
	// bukan diabaikan diam-diam.
	if filter.Status != "" && !service.ValidProjectStatuses[filter.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status hanya boleh 'draft' atau 'published'"})
		return
	}

	// Bila peminta adalah pemiliknya, draft miliknya ikut tampil
	if filter.Status == "draft" {
		filter.OwnerID = viewerID
	}

	projects, err := h.service.GetAllProjects(filter, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": projects})
}

// GetByID menampilkan detail project (publik; draft hanya untuk pemiliknya).
// GET /api/projects/:id
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	viewerID := getOptionalUserID(c)

	project, err := h.service.GetProjectByID(id, viewerID)
	if err != nil {
		mapProjectError(c, err, "Anda tidak memiliki izin untuk melihat project ini")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": project})
}

// Update memperbarui project milik user yang sedang login (terproteksi).
// PUT /api/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var input ProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}

	updates := &repository.Project{
		Title:       input.Title,
		Description: input.Description,
		TechStack:   input.TechStack,
		RepoURL:     input.RepoURL,
		DemoURL:     input.DemoURL,
		ImageURL:    input.ImageURL,
		Status:      input.Status,
	}

	updated, err := h.service.UpdateProject(id, userID.(string), updates, buildGallery(input.Gallery))
	if err != nil {
		mapProjectError(c, err, "Anda tidak memiliki izin untuk mengedit project ini")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project berhasil diperbarui",
		"data":    updated,
	})
}

// Delete menghapus project milik user yang sedang login (terproteksi).
// DELETE /api/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	projectID := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteProject(projectID, userID.(string))
	if err != nil {
		mapProjectError(c, err, "Anda tidak memiliki izin untuk menghapus project ini")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project berhasil dihapus secara permanen",
	})
}

// GetStatuses menampilkan daftar status project yang valid (draft/published).
// Dipakai frontend untuk mengisi pilihan status pada form tanpa hardcode.
// GET /api/projects/statuses (publik)
func (h *ProjectHandler) GetStatuses(c *gin.Context) {
	statuses := make([]string, 0, len(service.ValidProjectStatuses))
	for s := range service.ValidProjectStatuses {
		statuses = append(statuses, s)
	}
	sort.Strings(statuses)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Berhasil",
		"statuses": statuses,
	})
}
