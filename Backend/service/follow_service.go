package service

import (
	"campusconnect/repository"
	"errors"
)

type FollowService interface {
	ToggleFollow(followerID, followingID string) (following bool, err error)
	// GetUserWithFollowStatus mengembalikan data user publik beserta status
	// apakah viewer (viewerID, boleh kosong) mengikutinya.
	GetUserWithFollowStatus(userID, viewerID string) (*repository.User, bool, error)
}

type followService struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
}

func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository) FollowService {
	return &followService{followRepo: followRepo, userRepo: userRepo}
}

// GetUserWithFollowStatus mengembalikan data user publik + status apakah
// viewer sudah meng-follow user tersebut. Dipakai halaman profil publik agar
// tombol Follow selalu sinkron dengan server setelah refresh.
func (s *followService) GetUserWithFollowStatus(userID, viewerID string) (*repository.User, bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, false, err // repository.ErrNotFound bila tidak ada
	}

	following := false
	if viewerID != "" && viewerID != userID {
		_, err = s.followRepo.FindByFollowerAndFollowing(viewerID, userID)
		if err == nil {
			following = true
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, false, err
		}
	}

	return user, following, nil
}

// ToggleFollow membalik status follow: jika belum follow maka follow,
// jika sudah follow maka unfollow. Menolak self-follow di level service
// agar error DB constraint (chk_no_self_follow) tidak pernah terjadi.
func (s *followService) ToggleFollow(followerID, followingID string) (bool, error) {
	// 1. Tolak follow diri sendiri secara eksplisit (400, bukan 500)
	if followerID == followingID {
		return false, errors.New("anda tidak bisa follow diri sendiri")
	}

	// 2. Pastikan user target benar-benar ada (404 jika tidak)
	if _, err := s.userRepo.FindByID(followingID); err != nil {
		return false, err // Sudah berupa repository.ErrNotFound
	}

	// 3. Cek apakah sudah follow
	_, err := s.followRepo.FindByFollowerAndFollowing(followerID, followingID)

	if err == nil {
		// Sudah follow → unfollow
		if delErr := s.followRepo.Delete(followerID, followingID); delErr != nil {
			return false, delErr
		}
		return false, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return false, err
	}

	// 4. Belum follow → follow
	if createErr := s.followRepo.Create(&repository.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}); createErr != nil {
		return false, createErr
	}
	return true, nil
}
