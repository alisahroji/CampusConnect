package service

import (
	"campusconnect/repository"
	"errors"
)

// BookmarkService menangani logika bookmark untuk Project & Post (Minggu 6).
// Pola identik dengan LikeService: toggle (add/remove), validasi keberadaan
// target, dan sentinel error repository (ErrNotFound / ErrForbidden).
type BookmarkService interface {
	// ToggleProjectBookmark: bookmark bila belum ada, hapus bila sudah.
	// Mengembalikan status bookmarked terbaru.
	ToggleProjectBookmark(projectID, userID string) (bool, error)
	// GetProjectBookmarkStatus mengembalikan status bookmark user pada project.
	// userID kosong berarti anonymous (bookmarked = false).
	GetProjectBookmarkStatus(projectID, userID string) (bool, error)
	// ListProjectBookmarks mengembalikan seluruh project milik userID.
	ListProjectBookmarks(userID string) ([]repository.Project, error)

	TogglePostBookmark(postID, userID string) (bool, error)
	GetPostBookmarkStatus(postID, userID string) (bool, error)
	ListPostBookmarks(userID string) ([]repository.Post, error)
}

type bookmarkService struct {
	projectBookmarkRepo repository.ProjectBookmarkRepository
	postBookmarkRepo    repository.PostBookmarkRepository
	projectRepo         repository.ProjectRepository
	postRepo            repository.PostRepository
}

func NewBookmarkService(
	projectBookmarkRepo repository.ProjectBookmarkRepository,
	postBookmarkRepo repository.PostBookmarkRepository,
	projectRepo repository.ProjectRepository,
	postRepo repository.PostRepository,
) BookmarkService {
	return &bookmarkService{
		projectBookmarkRepo: projectBookmarkRepo,
		postBookmarkRepo:    postBookmarkRepo,
		projectRepo:         projectRepo,
		postRepo:            postRepo,
	}
}

// ToggleProjectBookmark membalik status bookmark project seorang user.
// Target harus ada (404 bila tidak). Duplicate dicegah oleh unique constraint
// uq_project_bookmarks_pair di level database, ditambah cek eksplisit di sini
// agar alur toggle tidak pernah menyentuh constraint tersebut.
func (s *bookmarkService) ToggleProjectBookmark(projectID, userID string) (bool, error) {
	// Target harus ada (draft ikut boleh di-bookmark oleh pemiliknya karena
	// FindByID di sini memanggil repo langsung; visibility draft ditangani
	// layer project saat menampilkan detail/list publik).
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		return false, err // repository.ErrNotFound bila tidak ada
	}

	_, err := s.projectBookmarkRepo.FindByUserAndProject(projectID, userID)
	if err == nil {
		// Sudah di-bookmark -> hapus (unbookmark)
		if delErr := s.projectBookmarkRepo.Delete(projectID, userID); delErr != nil {
			return false, delErr
		}
		return false, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return false, err
	}

	// Belum ada -> bookmark
	if createErr := s.projectBookmarkRepo.Create(&repository.ProjectBookmark{
		ProjectID: projectID,
		UserID:    userID,
	}); createErr != nil {
		return false, createErr
	}
	return true, nil
}

func (s *bookmarkService) GetProjectBookmarkStatus(projectID, userID string) (bool, error) {
	// Project harus ada (404 bila tidak) agar status tidak mengambang
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		return false, err
	}

	if userID == "" {
		return false, nil
	}
	_, err := s.projectBookmarkRepo.FindByUserAndProject(projectID, userID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (s *bookmarkService) ListProjectBookmarks(userID string) ([]repository.Project, error) {
	return s.projectBookmarkRepo.ListByUser(userID)
}

// TogglePostBookmark membalik status bookmark post seorang user.
func (s *bookmarkService) TogglePostBookmark(postID, userID string) (bool, error) {
	if _, err := s.postRepo.FindByID(postID, userID); err != nil {
		return false, err // repository.ErrNotFound bila tidak ada
	}

	_, err := s.postBookmarkRepo.FindByUserAndPost(postID, userID)
	if err == nil {
		if delErr := s.postBookmarkRepo.Delete(postID, userID); delErr != nil {
			return false, delErr
		}
		return false, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return false, err
	}

	if createErr := s.postBookmarkRepo.Create(&repository.PostBookmark{
		PostID: postID,
		UserID: userID,
	}); createErr != nil {
		return false, createErr
	}
	return true, nil
}

func (s *bookmarkService) GetPostBookmarkStatus(postID, userID string) (bool, error) {
	if _, err := s.postRepo.FindByID(postID, userID); err != nil {
		return false, err
	}

	if userID == "" {
		return false, nil
	}
	_, err := s.postBookmarkRepo.FindByUserAndPost(postID, userID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (s *bookmarkService) ListPostBookmarks(userID string) ([]repository.Post, error) {
	return s.postBookmarkRepo.ListByUser(userID)
}
