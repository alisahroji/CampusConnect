package service

import (
	"campusconnect/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// MOCK REPOSITORY
// =============================================================================

// MockProjectRepository meniru ProjectRepository agar test tidak menyentuh DB asli.
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(project *repository.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) FindByID(id string) (*repository.Project, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*repository.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProjectRepository) FindAll(filter repository.ProjectFilter) ([]repository.Project, error) {
	args := m.Called(filter)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProjectRepository) FindAllByUserID(userID string) ([]repository.Project, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProjectRepository) Update(project *repository.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockGalleryRepository meniru GalleryRepository.
type MockGalleryRepository struct {
	mock.Mock
}

func (m *MockGalleryRepository) CreateAll(galleries []repository.ProjectGallery) error {
	args := m.Called(galleries)
	return args.Error(0)
}

func (m *MockGalleryRepository) DeleteAllByProjectID(projectID string) error {
	args := m.Called(projectID)
	return args.Error(0)
}

// newProjectServiceWithMocks membuat service + kedua mock sekaligus.
func newProjectServiceWithMocks() (ProjectService, *MockProjectRepository, *MockGalleryRepository) {
	mockRepo := new(MockProjectRepository)
	mockGallery := new(MockGalleryRepository)
	svc := NewProjectService(mockRepo, mockGallery)
	return svc, mockRepo, mockGallery
}

// =============================================================================
// TEST CREATE
// =============================================================================

func TestCreateProject_Success(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	project := &repository.Project{
		Title:       "Smart Kampus",
		Description: "Aplikasi kampus pintar",
		Status:      "published",
	}

	mockRepo.On("Create", mock.MatchedBy(func(p *repository.Project) bool {
		return p.Title == "Smart Kampus" && p.Status == "published"
	})).Return(nil).Once()

	err := svc.CreateProject(project, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateProject_MissingFields(t *testing.T) {
	svc, _, _ := newProjectServiceWithMocks()

	// Judul kosong
	err := svc.CreateProject(&repository.Project{Description: "ada"}, nil)
	assert.Error(t, err)

	// Deskripsi kosong
	err = svc.CreateProject(&repository.Project{Title: "ada"}, nil)
	assert.Error(t, err)
}

func TestCreateProject_InvalidStatus(t *testing.T) {
	svc, _, _ := newProjectServiceWithMocks()

	err := svc.CreateProject(&repository.Project{
		Title:       "Judul",
		Description: "Deskripsi",
		Status:      "archived", // Tidak ada di whitelist
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestCreateProject_EmptyStatusDefaultsToPublished(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	project := &repository.Project{Title: "Judul", Description: "Deskripsi", Status: ""}

	mockRepo.On("Create", mock.MatchedBy(func(p *repository.Project) bool {
		return p.Status == "published" // Default terpasang
	})).Return(nil).Once()

	err := svc.CreateProject(project, nil)

	assert.NoError(t, err)
	assert.Equal(t, "published", project.Status)
	mockRepo.AssertExpectations(t)
}

func TestCreateProject_WithGallery(t *testing.T) {
	svc, mockRepo, mockGallery := newProjectServiceWithMocks()

	project := &repository.Project{Title: "Judul", Description: "Deskripsi", Status: ""}
	galleryInput := []repository.ProjectGallery{{ImageURL: "https://a.jpg"}, {ImageURL: "https://b.jpg"}}

	mockRepo.On("Create", mock.AnythingOfType("*repository.Project")).Return(nil).Once()
	mockGallery.On("CreateAll", mock.MatchedBy(func(g []repository.ProjectGallery) bool {
		// display_order harus deterministik sesuai urutan payload & project_id terisi
		return len(g) == 2 && g[0].DisplayOrder == 0 && g[1].DisplayOrder == 1 && g[0].ProjectID == project.ID
	})).Return(nil).Once()

	err := svc.CreateProject(project, galleryInput)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockGallery.AssertExpectations(t)
}

// =============================================================================
// TEST DRAFT VISIBILITY (GetByID & List)
// =============================================================================

func TestGetProjectByID_DraftHiddenFromOthers(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	draft := &repository.Project{ID: "p-1", UserID: "owner-1", Status: "draft"}
	mockRepo.On("FindByID", "p-1").Return(draft, nil) // dipanggil 3x untuk 3 kasus viewer

	// Viewer bukan pemilik → harus ErrNotFound (draft disembunyikan)
	_, err := svc.GetProjectByID("p-1", "other-user")
	assert.ErrorIs(t, err, repository.ErrNotFound)

	// Anonymous juga tidak boleh
	_, err = svc.GetProjectByID("p-1", "")
	assert.ErrorIs(t, err, repository.ErrNotFound)

	// Pemilik boleh melihat draft-nya sendiri
	found, err := svc.GetProjectByID("p-1", "owner-1")
	assert.NoError(t, err)
	assert.Equal(t, draft, found)

	mockRepo.AssertExpectations(t)
}

func TestGetProjectByID_PublishedVisibleToAll(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	published := &repository.Project{ID: "p-2", UserID: "owner-1", Status: "published"}
	mockRepo.On("FindByID", "p-2").Return(published, nil).Once()

	found, err := svc.GetProjectByID("p-2", "random-viewer")
	assert.NoError(t, err)
	assert.Equal(t, published, found)

	mockRepo.AssertExpectations(t)
}

func TestGetAllProjects_DraftQueryFromOtherUserReturnsEmpty(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	// Viewer "lain" minta list draft → service harus menolak TANPA menyentuh repo
	projects, err := svc.GetAllProjects(repository.ProjectFilter{Status: "draft", OwnerID: "other"}, "viewer-1")

	assert.NoError(t, err)
	assert.Empty(t, projects)
	mockRepo.AssertNotCalled(t, "FindAll", mock.Anything)
}

func TestGetAllProjects_PassesFilterToRepository(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	filter := repository.ProjectFilter{Status: "published", TechStack: "React"}
	expected := []repository.Project{{ID: "p-3", Status: "published"}}

	mockRepo.On("FindAll", filter).Return(expected, nil).Once()

	projects, err := svc.GetAllProjects(filter, "viewer-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, projects)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// TEST UPDATE (ownership + gallery sync)
// =============================================================================

func TestUpdateProject_Success(t *testing.T) {
	svc, mockRepo, mockGallery := newProjectServiceWithMocks()

	existing := &repository.Project{ID: "p-4", UserID: "owner-1", Status: "published"}
	updates := &repository.Project{Title: "Baru", Description: "Desc", Status: "draft"}
	galleryInput := []repository.ProjectGallery{{ImageURL: "https://x.jpg"}}

	mockRepo.On("FindByID", "p-4").Return(existing, nil).Once()
	mockRepo.On("Update", mock.MatchedBy(func(p *repository.Project) bool {
		return p.Title == "Baru" && p.Status == "draft"
	})).Return(nil).Once()
	mockGallery.On("DeleteAllByProjectID", "p-4").Return(nil).Once()
	mockGallery.On("CreateAll", mock.AnythingOfType("[]repository.ProjectGallery")).Return(nil).Once()
	// Reload setelah sync
	mockRepo.On("FindByID", "p-4").Return(existing, nil).Once()

	updated, err := svc.UpdateProject("p-4", "owner-1", updates, galleryInput)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	mockRepo.AssertExpectations(t)
	mockGallery.AssertExpectations(t)
}

func TestUpdateProject_ForbiddenForNonOwner(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	existing := &repository.Project{ID: "p-5", UserID: "owner-1", Status: "published"}
	mockRepo.On("FindByID", "p-5").Return(existing, nil).Once()

	_, err := svc.UpdateProject("p-5", "bukan-pemilik", &repository.Project{Title: "X", Description: "Y"}, nil)

	assert.ErrorIs(t, err, repository.ErrForbidden)
	// Tidak boleh sampai menyentuh Update/Delete galeri
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProject_InvalidStatus(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	existing := &repository.Project{ID: "p-6", UserID: "owner-1", Status: "published"}
	mockRepo.On("FindByID", "p-6").Return(existing, nil).Once()

	_, err := svc.UpdateProject("p-6", "owner-1", &repository.Project{
		Title: "X", Description: "Y", Status: "secret",
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProject_MissingFields(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	existing := &repository.Project{ID: "p-7", UserID: "owner-1", Status: "published"}
	mockRepo.On("FindByID", "p-7").Return(existing, nil).Once()

	_, err := svc.UpdateProject("p-7", "owner-1", &repository.Project{Title: "", Description: ""}, nil)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// TEST DELETE (ownership)
// =============================================================================

func TestDeleteProject_Success(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	existing := &repository.Project{ID: "p-8", UserID: "owner-1"}
	mockRepo.On("FindByID", "p-8").Return(existing, nil).Once()
	mockRepo.On("Delete", "p-8").Return(nil).Once()

	err := svc.DeleteProject("p-8", "owner-1")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProject_ForbiddenForNonOwner(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	existing := &repository.Project{ID: "p-9", UserID: "owner-1"}
	mockRepo.On("FindByID", "p-9").Return(existing, nil).Once()

	err := svc.DeleteProject("p-9", "orang-lain")

	assert.ErrorIs(t, err, repository.ErrForbidden)
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProject_NotFound(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	mockRepo.On("FindByID", "ghost").Return(nil, repository.ErrNotFound).Once()

	err := svc.DeleteProject("ghost", "owner-1")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// TEST GET MY PROJECTS
// =============================================================================

func TestGetMyProjects_ReturnsAllIncludingDraft(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	expected := []repository.Project{
		{ID: "p-a", UserID: "me", Status: "published"},
		{ID: "p-b", UserID: "me", Status: "draft"},
	}
	mockRepo.On("FindAllByUserID", "me").Return(expected, nil).Once()

	projects, err := svc.GetMyProjects("me")

	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetMyProjects_RepoErrorPropagates(t *testing.T) {
	svc, mockRepo, _ := newProjectServiceWithMocks()

	mockRepo.On("FindAllByUserID", "me").Return(nil, errors.New("db down")).Once()

	_, err := svc.GetMyProjects("me")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
