package service

import (
	"campusconnect/repository"
	"errors"
)

type ProjectService interface {
	CreateProject(project *repository.Project) error
	GetProjectByID(id string) (*repository.Project, error)
	GetAllProjects(status string) ([]repository.Project, error)
	UpdateProject(project *repository.Project) error
	DeleteProject(id string, userID string) error
}

type projectService struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) CreateProject(project *repository.Project) error {
	// Validasi dasar
	if project.Title == "" || project.Description == "" {
		return errors.New("judul dan deskripsi project wajib diisi")
	}
	return s.repo.Create(project)
}

func (s *projectService) GetProjectByID(id string) (*repository.Project, error) {
	return s.repo.FindByID(id)
}

func (s *projectService) GetAllProjects(status string) ([]repository.Project, error) {
	return s.repo.FindAll(status)
}

func (s *projectService) UpdateProject(project *repository.Project) error {
	return s.repo.Update(project)
}

func (s *projectService) DeleteProject(projectID string, userID string) error {
	// 1. Cari projectnya dulu
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return err // Sudah berupa repository.ErrNotFound
	}

	// 2. Pastikan yang menghapus adalah pemiliknya (atau admin)
	// Saat ini kita cek kepemilikan.
	if project.UserID != userID {
		return repository.ErrForbidden
	}

	return s.repo.Delete(projectID)
}