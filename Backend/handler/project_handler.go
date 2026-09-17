package handler

import (
	"campusconnect/repository"
	"campusconnect/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service service.ProjectService
}

func NewProjectHandler(service service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// Input struct untuk membuat/update project
type ProjectInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	TechStack   string `json:"tech_stack"`
	RepoURL     string `json:"repo_url"`
	DemoURL     string `json:"demo_url"`
	ImageURL    string `json:"image_url"`
	Status      string `json:"status"`
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var input ProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}

	// Ambil ID User dari token JWT (dari middleware RequireAuth)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
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

	if project.Status == "" {
		project.Status = "published"
	}

	if err := h.service.CreateProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project berhasil dibuat",
		"data":    project,
	})
}

func (h *ProjectHandler) GetAll(c *gin.Context) {
	status := c.Query("status") // Bisa dikosongkan untuk mengambil semua
	projects, err := h.service.GetAllProjects(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil",
		"data":    projects,
	})
}

// ... (Kode sebelumnya: Create dan GetAll)

func (h *ProjectHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	project, err := h.service.GetProjectByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil",
		"data":    project,
	})
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// 1. Pastikan project-nya ada dan dimiliki oleh user yang sedang login
	existingProject, err := h.service.GetProjectByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil project"})
		return
	}

	userID, _ := c.Get("userID")
	if existingProject.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk mengedit project ini"})
		return
	}

	// 2. Tangkap input perubahan dari user
	var input ProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "details": err.Error()})
		return
	}

	// 3. Timpa data lama dengan data baru
	existingProject.Title = input.Title
	existingProject.Description = input.Description
	existingProject.TechStack = input.TechStack
	existingProject.RepoURL = input.RepoURL
	existingProject.DemoURL = input.DemoURL
	existingProject.ImageURL = input.ImageURL
	existingProject.Status = input.Status

	if existingProject.Status == "" {
		existingProject.Status = "published"
	}

	// 4. Simpan ke database
	if err := h.service.UpdateProject(existingProject); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project berhasil diperbarui",
		"data":    existingProject,
	})
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	projectID := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteProject(projectID, userID.(string))
	if err != nil {
		// Map sentinel errors ke status HTTP yang tepat
		if errors.Is(err, repository.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk menghapus project ini"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project berhasil dihapus secara permanen",
	})
}