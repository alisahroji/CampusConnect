package service

import (
	"campusconnect/repository"
	"errors"
)

type LikeService interface {
	ToggleLike(projectID, userID string) (liked bool, likeCount int64, err error)
	GetLikeCount(projectID string) (int64, error)
	// GetLikeInfo mengembalikan jumlah like + status liked user tertentu.
	// userID kosong berarti anonymous (liked = false).
	GetLikeInfo(projectID, userID string) (likeCount int64, liked bool, err error)
	// SetNotifier opsional (Minggu 6): membuat notification ke pemilik project.
	SetNotifier(n Notifier)
}

type likeService struct {
	likeRepo    repository.LikeRepository
	projectRepo repository.ProjectRepository
	notifier    Notifier // opsional (Minggu 6): dipasang lewat SetNotifier
}

func NewLikeService(likeRepo repository.LikeRepository, projectRepo repository.ProjectRepository) LikeService {
	return &likeService{likeRepo: likeRepo, projectRepo: projectRepo}
}

// SetNotifier memasang Notifier (Minggu 6 Hari 2) tanpa mengubah constructor
// agar seluruh wiring & test existing tidak berubah. Nil nil = tanpa notifikasi.
func (s *likeService) SetNotifier(n Notifier) { s.notifier = n }

// ToggleLike membalik status like: jika sudah like maka di-unlike, sebaliknya di-like.
func (s *likeService) ToggleLike(projectID, userID string) (bool, int64, error) {
	// 1. Pastikan project-nya benar-benar ada
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return false, 0, err // Sudah berupa repository.ErrNotFound
	}

	// 2. Cek apakah user sudah like project ini
	_, err = s.likeRepo.FindByUserAndProject(projectID, userID)
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
		// Minggu 6: beri tahu pemilik project (skip bila user == pemilik)
		if s.notifier != nil {
			s.notifier.Notify(project.UserID, userID, repository.NotifTypeLikeProject, projectID)
		}
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

// GetLikeInfo mengembalikan jumlah like sebuah project beserta status apakah
// user (userID) saat ini sudah me-like-nya. Dipakai frontend agar ikon like
// tetap "liked" setelah refresh (BUG-1: state like tidak persist).
func (s *likeService) GetLikeInfo(projectID, userID string) (int64, bool, error) {
	// Pastikan project-nya ada (404 bila tidak)
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		return 0, false, err
	}

	count, err := s.likeRepo.CountByProject(projectID)
	if err != nil {
		return 0, false, err
	}

	liked := false
	if userID != "" {
		_, err = s.likeRepo.FindByUserAndProject(projectID, userID)
		if err == nil {
			liked = true
		} else if !errors.Is(err, repository.ErrNotFound) {
			return 0, false, err
		}
	}

	return count, liked, nil
}
