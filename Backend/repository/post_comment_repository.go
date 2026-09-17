package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema PostComment untuk database (Minggu 5 - interaksi post)
type PostComment struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    string    `gorm:"type:uuid;not null;index" json:"post_id"` // Relasi ke post yang dikomentari
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"` // Relasi ke penulis komentar
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relasi ke tabel User (penulis komentar)
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// Interface standar Clean Architecture
type PostCommentRepository interface {
	Create(comment *PostComment) error
	FindByID(id string) (*PostComment, error)
	FindByPostID(postID string) ([]PostComment, error)
	Delete(id string) error
}

type postCommentRepository struct {
	db *gorm.DB
}

func NewPostCommentRepository(db *gorm.DB) PostCommentRepository {
	return &postCommentRepository{db: db}
}

func (r *postCommentRepository) Create(comment *PostComment) error {
	if err := r.db.Create(comment).Error; err != nil {
		return err
	}
	// Muat relasi User agar response langsung berisi data penulis komentar
	return r.db.Preload("User").First(comment, "id = ?", comment.ID).Error
}

func (r *postCommentRepository) FindByID(id string) (*PostComment, error) {
	var comment PostComment
	err := r.db.Preload("User").Where("id = ?", id).First(&comment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *postCommentRepository) FindByPostID(postID string) ([]PostComment, error) {
	var comments []PostComment
	err := r.db.Preload("User").Where("post_id = ?", postID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *postCommentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&PostComment{}).Error
}
