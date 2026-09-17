package service

import (
	"campusconnect/repository"
)

const (
	defaultFeedLimit = 10
	maxFeedLimit     = 50
)

type FeedService interface {
	GetFeed(userID, cursor string, limit int) ([]repository.Post, string, error)
}

type feedService struct {
	followRepo repository.FollowRepository
	postRepo   repository.PostRepository
}

func NewFeedService(followRepo repository.FollowRepository, postRepo repository.PostRepository) FeedService {
	return &feedService{followRepo: followRepo, postRepo: postRepo}
}

// GetFeed mengembalikan post HANYA dari user yang di-follow userID,
// DITAMBAH post milik userID sendiri, diurutkan dari yang terbaru,
// dengan cursor-based pagination (keyset) untuk infinite scrolling.
// Berbeda dengan GET /api/posts yang bersifat global (explore).
func (s *feedService) GetFeed(userID, cursor string, limit int) ([]repository.Post, string, error) {
	if limit <= 0 {
		limit = defaultFeedLimit
	}
	if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	// 1. Ambil daftar ID user yang di-follow
	followingIDs, err := s.followRepo.FindFollowingIDs(userID)
	if err != nil {
		return nil, "", err
	}

	// 2. Gabungkan dengan post milik sendiri
	userIDs := append(followingIDs, userID)

	// 3. Ambil post dari kumpulan user tersebut dengan cursor
	// viewer = user yang meminta feed, agar liked_by_me terisi benar.
	return s.postRepo.FindByUserIDs(userIDs, cursor, limit, userID)
}
