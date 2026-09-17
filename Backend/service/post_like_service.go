package service

import (
	"campusconnect/repository"
	"errors"
)

type PostLikeService interface {
	TogglePostLike(postID, userID string) (liked bool, likeCount int64, err error)
	GetPostLikeCount(postID string) (int64, error)
}

type postLikeService struct {
	postLikeRepo repository.PostLikeRepository
	postRepo     repository.PostRepository
}

func NewPostLikeService(postLikeRepo repository.PostLikeRepository, postRepo repository.PostRepository) PostLikeService {
	return &postLikeService{postLikeRepo: postLikeRepo, postRepo: postRepo}
}

// TogglePostLike membalik status like pada sebuah post: jika sudah like maka
// di-unlike, jika belum maka di-like. Mirip dengan toggle like pada project.
func (s *postLikeService) TogglePostLike(postID, userID string) (bool, int64, error) {
	// 1. Pastikan post-nya benar-benar ada
	if _, err := s.postRepo.FindByID(postID); err != nil {
		return false, 0, err // Sudah berupa repository.ErrNotFound
	}

	// 2. Cek apakah user sudah like post ini
	_, err := s.postLikeRepo.FindByUserAndPost(postID, userID)
	liked := false

	if err == nil {
		// Sudah like → hapus (unlike)
		if delErr := s.postLikeRepo.Delete(postID, userID); delErr != nil {
			return false, 0, delErr
		}
		liked = false
	} else if errors.Is(err, repository.ErrNotFound) {
		// Belum like → tambahkan (like)
		if createErr := s.postLikeRepo.Create(&repository.PostLike{PostID: postID, UserID: userID}); createErr != nil {
			return false, 0, createErr
		}
		liked = true
	} else {
		return false, 0, err
	}

	// 3. Hitung total like terbaru
	count, countErr := s.postLikeRepo.CountByPost(postID)
	if countErr != nil {
		return liked, 0, countErr
	}

	return liked, count, nil
}

func (s *postLikeService) GetPostLikeCount(postID string) (int64, error) {
	if _, err := s.postRepo.FindByID(postID); err != nil {
		return 0, err // Sudah berupa repository.ErrNotFound
	}
	return s.postLikeRepo.CountByPost(postID)
}
