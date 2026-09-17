package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema Follow untuk database (Minggu 5 - Social Graph)
type Follow struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FollowerID  string    `gorm:"type:uuid;not null;index" json:"follower_id"`  // Yang memberi follow
	FollowingID string    `gorm:"type:uuid;not null;index" json:"following_id"` // Yang di-follow
	CreatedAt   time.Time `json:"created_at"`
}

// Interface standar Clean Architecture
type FollowRepository interface {
	FindByFollowerAndFollowing(followerID, followingID string) (*Follow, error)
	Create(follow *Follow) error
	Delete(followerID, followingID string) error
	// FindFollowingIDs mengembalikan daftar ID user yang di-follow oleh seorang user.
	// Dipakai oleh query feed berbasis following.
	FindFollowingIDs(followerID string) ([]string, error)
}

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) FindByFollowerAndFollowing(followerID, followingID string) (*Follow, error) {
	var follow Follow
	err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *followRepository) Create(follow *Follow) error {
	return r.db.Create(follow).Error
}

func (r *followRepository) Delete(followerID, followingID string) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&Follow{}).Error
}

func (r *followRepository) FindFollowingIDs(followerID string) ([]string, error) {
	var ids []string
	err := r.db.Model(&Follow{}).Where("follower_id = ?", followerID).Pluck("following_id", &ids).Error
	return ids, err
}
