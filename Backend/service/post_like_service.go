package service

import (
	"campusconnect/repository"
	"errors"
)

type PostLikeService interface {
	TogglePostLike(postID, userID string) (liked bool, likeCount int64, err error)
	GetPostLikeCount(postID string) (int64, error)
	// GetPostLikeInfo mengembalikan jumlah like + status liked user tertentu.
	// userID kosong berarti anonymous (liked = false).
	GetPostLikeInfo(postID, userID string) (likeCount int64, liked bool, err error)
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
	if _, err := s.postRepo.FindByID(postID, ""); err != nil {
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
	if _, err := s.postRepo.FindByID(postID, ""); err != nil {
		return 0, err // Sudah berupa repository.ErrNotFound
	}
	return s.postLikeRepo.CountByPost(postID)
}

// GetPostLikeInfo mengembalikan jumlah like sebuah post beserta status apakah
// user (userID) saat ini sudah me-like-nya. Dipakai frontend agar ikon like
// post tetap "liked" setelah refresh (BUG-5A). userID kosong = anonymous.
func (s *postLikeService) GetPostLikeInfo(postID, userID string) (int64, bool, error) {
	// Pastikan post-nya ada (404 bila tidak)
	if _, err := s.postRepo.FindByID(postID, ""); err != nil {
		return 0, false, err
	}

	count, err := s.postLikeRepo.CountByPost(postID)
	if err != nil {
		return 0, false, err
	}

	liked := false
	if userID != "" {
		_, err = s.postLikeRepo.FindByUserAndPost(postID, userID)
		if err == nil {
			liked = true
		} else if !errors.Is(err, repository.ErrNotFound) {
			return 0, false, err
		}
	}

	return count, liked, nil
}
