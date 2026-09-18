package service

import (
	"campusconnect/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// MOCK REPOSITORY BOOKMARK (Minggu 6 Hari 1)
// MockProjectRepository & MockPostRepository di-reuse dari test file lain
// (project_service_test.go & post_service_test.go, package yang sama).
// =============================================================================

type MockProjectBookmarkRepository struct {
	mock.Mock
}

func (m *MockProjectBookmarkRepository) FindByUserAndProject(projectID, userID string) (*repository.ProjectBookmark, error) {
	args := m.Called(projectID, userID)
	bookmark, _ := args.Get(0).(*repository.ProjectBookmark)
	return bookmark, args.Error(1)
}

func (m *MockProjectBookmarkRepository) Create(bookmark *repository.ProjectBookmark) error {
	args := m.Called(bookmark)
	return args.Error(0)
}

func (m *MockProjectBookmarkRepository) Delete(projectID, userID string) error {
	args := m.Called(projectID, userID)
	return args.Error(0)
}

func (m *MockProjectBookmarkRepository) ListByUser(userID string) ([]repository.Project, error) {
	args := m.Called(userID)
	projects, _ := args.Get(0).([]repository.Project)
	return projects, args.Error(1)
}

type MockPostBookmarkRepository struct {
	mock.Mock
}

func (m *MockPostBookmarkRepository) FindByUserAndPost(postID, userID string) (*repository.PostBookmark, error) {
	args := m.Called(postID, userID)
	bookmark, _ := args.Get(0).(*repository.PostBookmark)
	return bookmark, args.Error(1)
}

func (m *MockPostBookmarkRepository) Create(bookmark *repository.PostBookmark) error {
	args := m.Called(bookmark)
	return args.Error(0)
}

func (m *MockPostBookmarkRepository) Delete(postID, userID string) error {
	args := m.Called(postID, userID)
	return args.Error(0)
}

func (m *MockPostBookmarkRepository) ListByUser(userID string) ([]repository.Post, error) {
	args := m.Called(userID)
	posts, _ := args.Get(0).([]repository.Post)
	return posts, args.Error(1)
}

func newBookmarkTestService() (BookmarkService, *MockProjectBookmarkRepository, *MockPostBookmarkRepository, *MockProjectRepository, *MockPostRepository) {
	pbRepo := new(MockProjectBookmarkRepository)
	poRepo := new(MockPostBookmarkRepository)
	pRepo := new(MockProjectRepository)
	poMainRepo := new(MockPostRepository)
	svc := NewBookmarkService(pbRepo, poRepo, pRepo, poMainRepo)
	return svc, pbRepo, poRepo, pRepo, poMainRepo
}

// =============================================================================
// PROJECT BOOKMARK
// =============================================================================

func TestToggleProjectBookmark_Add(t *testing.T) {
	svc, pbRepo, _, pRepo, _ := newBookmarkTestService()

	pRepo.On("FindByID", "proj-1").Return(&repository.Project{ID: "proj-1"}, nil)
	// Belum pernah bookmark -> ErrNotFound
	pbRepo.On("FindByUserAndProject", "proj-1", "user-A").Return(nil, repository.ErrNotFound)
	pbRepo.On("Create", mock.MatchedBy(func(b *repository.ProjectBookmark) bool {
		return b.ProjectID == "proj-1" && b.UserID == "user-A"
	})).Return(nil)

	bookmarked, err := svc.ToggleProjectBookmark("proj-1", "user-A")

	assert.NoError(t, err)
	assert.True(t, bookmarked)
	pRepo.AssertExpectations(t)
	pbRepo.AssertExpectations(t)
}

func TestToggleProjectBookmark_Remove(t *testing.T) {
	svc, pbRepo, _, pRepo, _ := newBookmarkTestService()

	pRepo.On("FindByID", "proj-1").Return(&repository.Project{ID: "proj-1"}, nil)
	// Sudah ada bookmark -> dihapus
	pbRepo.On("FindByUserAndProject", "proj-1", "user-A").Return(&repository.ProjectBookmark{ProjectID: "proj-1", UserID: "user-A"}, nil)
	pbRepo.On("Delete", "proj-1", "user-A").Return(nil)

	bookmarked, err := svc.ToggleProjectBookmark("proj-1", "user-A")

	assert.NoError(t, err)
	assert.False(t, bookmarked)
	pbRepo.AssertExpectations(t)
}

func TestToggleProjectBookmark_ProjectMissing(t *testing.T) {
	svc, pbRepo, _, pRepo, _ := newBookmarkTestService()

	pRepo.On("FindByID", "tidak-ada").Return(nil, repository.ErrNotFound)

	bookmarked, err := svc.ToggleProjectBookmark("tidak-ada", "user-A")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.False(t, bookmarked)
	pbRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// Bookmark milik user lain tidak boleh memengaruhi status user ini
// (isolasi antar user pada pencarian pair user+project).
func TestGetProjectBookmarkStatus_IsolationBetweenUsers(t *testing.T) {
	svc, pbRepo, _, pRepo, _ := newBookmarkTestService()

	pRepo.On("FindByID", "proj-1").Return(&repository.Project{ID: "proj-1"}, nil).Twice()

	// user-A punya bookmark
	pbRepo.On("FindByUserAndProject", "proj-1", "user-A").Return(&repository.ProjectBookmark{}, nil)
	// user-B TIDAK punya bookmark untuk project yang sama
	pbRepo.On("FindByUserAndProject", "proj-1", "user-B").Return(nil, repository.ErrNotFound)

	bookmarkedA, errA := svc.GetProjectBookmarkStatus("proj-1", "user-A")
	bookmarkedB, errB := svc.GetProjectBookmarkStatus("proj-1", "user-B")

	assert.NoError(t, errA)
	assert.True(t, bookmarkedA)
	assert.NoError(t, errB)
	assert.False(t, bookmarkedB)
}

func TestGetProjectBookmarkStatus_Anonymous(t *testing.T) {
	svc, pbRepo, _, pRepo, _ := newBookmarkTestService()

	pRepo.On("FindByID", "proj-1").Return(&repository.Project{ID: "proj-1"}, nil)

	bookmarked, err := svc.GetProjectBookmarkStatus("proj-1", "")

	assert.NoError(t, err)
	assert.False(t, bookmarked)
	// Anonymous tidak boleh menyentuh tabel bookmark
	pbRepo.AssertNotCalled(t, "FindByUserAndProject", mock.Anything, mock.Anything)
}

func TestGetProjectBookmarkStatus_ProjectMissing(t *testing.T) {
	svc, _, _, pRepo, _ := newBookmarkTestService()

	pRepo.On("FindByID", "tidak-ada").Return(nil, repository.ErrNotFound)

	bookmarked, err := svc.GetProjectBookmarkStatus("tidak-ada", "user-A")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.False(t, bookmarked)
}

// =============================================================================
// POST BOOKMARK
// =============================================================================

func TestTogglePostBookmark_AddAndRemove(t *testing.T) {
	svc, _, poRepo, _, poMainRepo := newBookmarkTestService()

	poMainRepo.On("FindByID", "post-1", "user-A").Return(&repository.Post{ID: "post-1"}, nil)

	// Fase 1: add
	poRepo.On("FindByUserAndPost", "post-1", "user-A").Return(nil, repository.ErrNotFound).Once()
	poRepo.On("Create", mock.MatchedBy(func(b *repository.PostBookmark) bool {
		return b.PostID == "post-1" && b.UserID == "user-A"
	})).Return(nil).Once()

	bookmarked, err := svc.TogglePostBookmark("post-1", "user-A")
	assert.NoError(t, err)
	assert.True(t, bookmarked)

	// Fase 2: remove
	poRepo.On("FindByUserAndPost", "post-1", "user-A").Return(&repository.PostBookmark{PostID: "post-1", UserID: "user-A"}, nil).Once()
	poRepo.On("Delete", "post-1", "user-A").Return(nil).Once()

	bookmarked, err = svc.TogglePostBookmark("post-1", "user-A")
	assert.NoError(t, err)
	assert.False(t, bookmarked)

	poMainRepo.AssertExpectations(t)
	poRepo.AssertExpectations(t)
}

func TestTogglePostBookmark_PostMissing(t *testing.T) {
	svc, _, _, _, poMainRepo := newBookmarkTestService()

	poMainRepo.On("FindByID", "tidak-ada", "user-A").Return(nil, repository.ErrNotFound)

	bookmarked, err := svc.TogglePostBookmark("tidak-ada", "user-A")

	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.False(t, bookmarked)
}

func TestListProjectBookmarks(t *testing.T) {
	svc, pbRepo, _, _, _ := newBookmarkTestService()

	expected := []repository.Project{{ID: "proj-1"}, {ID: "proj-2"}}
	pbRepo.On("ListByUser", "user-A").Return(expected, nil)

	projects, err := svc.ListProjectBookmarks("user-A")

	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "proj-2", projects[1].ID)
}

func TestListPostBookmarks(t *testing.T) {
	svc, _, poRepo, _, _ := newBookmarkTestService()

	expected := []repository.Post{{ID: "post-1"}}
	poRepo.On("ListByUser", "user-A").Return(expected, nil)

	posts, err := svc.ListPostBookmarks("user-A")

	assert.NoError(t, err)
	assert.Len(t, posts, 1)
}

// Service tidak boleh mengubah error selain ErrNotFound dari layer di bawahnya
// (misalnya kegagalan DB tetap diteruskan agar handler memetakan ke 500).
func TestToggleProjectBookmark_UnexpectedErrorPropagates(t *testing.T) {
	svc, pbRepo, _, pRepo, _ := newBookmarkTestService()

	dbErr := errors.New("db down")
	pRepo.On("FindByID", "proj-1").Return(&repository.Project{ID: "proj-1"}, nil)
	pbRepo.On("FindByUserAndProject", "proj-1", "user-A").Return(nil, dbErr)

	bookmarked, err := svc.ToggleProjectBookmark("proj-1", "user-A")

	assert.ErrorIs(t, err, dbErr)
	assert.False(t, bookmarked)
}
