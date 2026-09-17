package service

import (
	"campusconnect/repository"
	"errors"
)

// Nilai status yang diizinkan untuk project (Whitelist).
// Diekspor agar handler bisa menampilkan daftar ini di endpoint /statuses
// tanpa menduplikasi sumber kebenaran.
var ValidProjectStatuses = map[string]bool{
	"draft":     true,
	"published": true,
}

// normalizeStatus memvalidasi status input: kosong berarti default "published",
// selain itu harus termasuk whitelist validProjectStatuses.
func normalizeStatus(status string) (string, error) {
	if status == "" {
		return "published", nil
	}
	if !ValidProjectStatuses[status] {
		return "", errors.New("status hanya boleh 'draft' atau 'published'")
	}
	return status, nil
}

type ProjectService interface {
	CreateProject(project *repository.Project, gallery []repository.ProjectGallery) error
	GetProjectByID(id string, viewerID string) (*repository.Project, error)
	GetAllProjects(filter repository.ProjectFilter, viewerID string) ([]repository.Project, error)
	GetMyProjects(userID string) ([]repository.Project, error)
	UpdateProject(projectID string, userID string, updates *repository.Project, gallery []repository.ProjectGallery) (*repository.Project, error)
	DeleteProject(id string, userID string) error
}

type projectService struct {
	repo        repository.ProjectRepository
	galleryRepo repository.GalleryRepository
}

func NewProjectService(repo repository.ProjectRepository, galleryRepo repository.GalleryRepository) ProjectService {
	return &projectService{repo: repo, galleryRepo: galleryRepo}
}

func (s *projectService) CreateProject(project *repository.Project, gallery []repository.ProjectGallery) error {
	// Validasi dasar
	if project.Title == "" || project.Description == "" {
		return errors.New("judul dan deskripsi project wajib diisi")
	}

	status, err := normalizeStatus(project.Status)
	if err != nil {
		return err
	}
	project.Status = status

	// Simpan project dulu agar ID tersedia untuk row galeri
	if err := s.repo.Create(project); err != nil {
		return err
	}

	// Simpan galeri (jika ada) dengan display_order sesuai urutan payload
	if len(gallery) > 0 {
		for i := range gallery {
			gallery[i].ProjectID = project.ID
			gallery[i].DisplayOrder = i
		}
		return s.galleryRepo.CreateAll(gallery)
	}

	return nil
}

// GetProjectByID mengambil detail project.
// Aturan visibility: draft hanya boleh dilihat oleh pemiliknya (viewerID == UserID).
// viewerID kosong berarti anonymous visitor.
func (s *projectService) GetProjectByID(id string, viewerID string) (*repository.Project, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err // Sudah berupa repository.ErrNotFound
	}

	if project.Status == "draft" && project.UserID != viewerID {
		return nil, repository.ErrNotFound // Draft disembunyikan, diperlakukan seperti tidak ada
	}

	return project, nil
}

// GetAllProjects menampilkan list project (publik/explore).
// Aturan visibility:
//   - Query draft eksplisit hanya dijawab bila pemiliknya yang meminta.
//   - List publik (tanpa owner) dipaksa published-only agar draft tidak
//     pernah bocor, termasuk saat dipanggil tanpa parameter status.
func (s *projectService) GetAllProjects(filter repository.ProjectFilter, viewerID string) ([]repository.Project, error) {
	// Query draft eksplisit hanya dijawab bila pemiliknya yang meminta
	// (anonymous tidak boleh melihat draft siapa pun).
	if filter.Status == "draft" && (filter.OwnerID == "" || filter.OwnerID != viewerID) {
		return []repository.Project{}, nil
	}
	// List publik tanpa scope owner dipaksa published-only agar draft
	// tidak pernah bocor, termasuk saat dipanggil tanpa parameter status.
	if filter.OwnerID == "" && filter.Status != "published" {
		filter.Status = "published"
	}
	return s.repo.FindAll(filter)
}

// GetMyProjects mengambil seluruh project milik user yang sedang login (termasuk draft).
func (s *projectService) GetMyProjects(userID string) ([]repository.Project, error) {
	return s.repo.FindAllByUserID(userID)
}

// UpdateProject memperbarui project milik userID. Ownership divalidasi di service
// agar terproteksi untuk semua pemanggil. Gallery di-sync dengan strategi
// "hapus semua lalu insert ulang" agar display_order selalu deterministik.
func (s *projectService) UpdateProject(projectID string, userID string, updates *repository.Project, gallery []repository.ProjectGallery) (*repository.Project, error) {
	// 1. Pastikan project ada
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return nil, err // repository.ErrNotFound
	}

	// 2. Ownership check
	if project.UserID != userID {
		return nil, repository.ErrForbidden
	}

	// 3. Validasi & terapkan perubahan field
	if updates.Title == "" || updates.Description == "" {
		return nil, errors.New("judul dan deskripsi project wajib diisi")
	}
	status, err := normalizeStatus(updates.Status)
	if err != nil {
		return nil, err
	}

	project.Title = updates.Title
	project.Description = updates.Description
	project.TechStack = updates.TechStack
	project.RepoURL = updates.RepoURL
	project.DemoURL = updates.DemoURL
	project.ImageURL = updates.ImageURL
	project.Status = status

	// 4. Simpan perubahan field utama
	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	// 5. Sync galeri (hapus semua -> insert ulang sesuai payload)
	if err := s.galleryRepo.DeleteAllByProjectID(projectID); err != nil {
		return nil, err
	}
	if len(gallery) > 0 {
		for i := range gallery {
			gallery[i].ProjectID = projectID
			gallery[i].DisplayOrder = i
		}
		if err := s.galleryRepo.CreateAll(gallery); err != nil {
			return nil, err
		}
	}

	// 6. Kembalikan data terbaru (beserta gallery hasil sync)
	return s.repo.FindByID(projectID)
}

func (s *projectService) DeleteProject(projectID string, userID string) error {
	// 1. Cari projectnya dulu
	project, err := s.repo.FindByID(projectID)
	if err != nil {
		return err // Sudah berupa repository.ErrNotFound
	}

	// 2. Pastikan yang menghapus adalah pemiliknya
	if project.UserID != userID {
		return repository.ErrForbidden
	}

	return s.repo.Delete(projectID)
}
