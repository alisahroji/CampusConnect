package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema Like untuk database
type Like struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID string    `gorm:"type:uuid;not null;index" json:"project_id"` // Relasi ke project yang di-like
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`    // Relasi ke user pemberi like
	CreatedAt time.Time `json:"created_at"`
}

// Interface standar Clean Architecture
type LikeRepository interface {
	FindByUserAndProject(projectID, userID string) (*Like, error)
	Create(like *Like) error
	Delete(projectID, userID string) error
	CountByProject(projectID string) (int64, error)
}

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) FindByUserAndProject(projectID, userID string) (*Like, error) {
	var like Like
	err := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&like).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *likeRepository) Create(like *Like) error {
	return r.db.Create(like).Error
}

func (r *likeRepository) Delete(projectID, userID string) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&Like{}).Error
}

func (r *likeRepository) CountByProject(projectID string) (int64, error) {
	var count int64
	err := r.db.Model(&Like{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}
