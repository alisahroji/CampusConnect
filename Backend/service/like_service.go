package service

import (
	"campusconnect/repository"
	"errors"
)

type LikeService interface {
	ToggleLike(projectID, userID string) (liked bool, likeCount int64, err error)
	GetLikeCount(projectID string) (int64, error)
}

type likeService struct {
	likeRepo    repository.LikeRepository
	projectRepo repository.ProjectRepository
}

func NewLikeService(likeRepo repository.LikeRepository, projectRepo repository.ProjectRepository) LikeService {
	return &likeService{likeRepo: likeRepo, projectRepo: projectRepo}
}

// ToggleLike membalik status like: jika sudah like maka di-unlike, sebaliknya di-like.
func (s *likeService) ToggleLike(projectID, userID string) (bool, int64, error) {
	// 1. Pastikan project-nya benar-benar ada
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		return false, 0, err // Sudah berupa repository.ErrNotFound
	}

	// 2. Cek apakah user sudah like project ini
	_, err := s.likeRepo.FindByUserAndProject(projectID, userID)
	liked := false

	if err == nil {
		// Sudah like → hapus (unlike)
		if delErr := s.likeRepo.Delete(projectID, userID); delErr != nil {
			return false, 0, delErr
		}
		liked = false
	} else if errors.Is(err, repository.ErrNotFound) {
		// Belum like → tambahkan (like)
		if createErr := s.likeRepo.Create(&repository.Like{ProjectID: projectID, UserID: userID}); createErr != nil {
			return false, 0, createErr
		}
		liked = true
	} else {
		return false, 0, err
	}

	// 3. Hitung total like terbaru
	count, countErr := s.likeRepo.CountByProject(projectID)
	if countErr != nil {
		return liked, 0, countErr
	}

	return liked, count, nil
}

func (s *likeService) GetLikeCount(projectID string) (int64, error) {
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		return 0, err // Sudah berupa repository.ErrNotFound
	}
	return s.likeRepo.CountByProject(projectID)
}
