package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema PostLike untuk database (Minggu 5 - interaksi post)
type PostLike struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"` // Relasi ke post yang di-like
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"` // Relasi ke user pemberi like
	CreatedAt time.Time `json:"created_at"`
}

// Interface standar Clean Architecture
type PostLikeRepository interface {
	FindByUserAndPost(postID, userID string) (*PostLike, error)
	Create(like *PostLike) error
	Delete(postID, userID string) error
	CountByPost(postID string) (int64, error)
}

type postLikeRepository struct {
	db *gorm.DB
}

func NewPostLikeRepository(db *gorm.DB) PostLikeRepository {
	return &postLikeRepository{db: db}
}

func (r *postLikeRepository) FindByUserAndPost(postID, userID string) (*PostLike, error) {
	var like PostLike
	err := r.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *postLikeRepository) Create(like *PostLike) error {
	return r.db.Create(like).Error
}

func (r *postLikeRepository) Delete(postID, userID string) error {
	return r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostLike{}).Error
}

func (r *postLikeRepository) CountByPost(postID string) (int64, error) {
	var count int64
	err := r.db.Model(&PostLike{}).Where("post_id = ?", postID).Count(&count).Error
	return count, err
}
