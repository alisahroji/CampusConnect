package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema Comment untuk database
type Comment struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID string    `gorm:"type:uuid;not null;index" json:"project_id"` // Relasi ke project yang dikomentari
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`    // Relasi ke penulis komentar
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relasi ke tabel User (penulis komentar)
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// Interface standar Clean Architecture
type CommentRepository interface {
	Create(comment *Comment) error
	FindByID(id string) (*Comment, error)
	FindByProjectID(projectID string) ([]Comment, error)
	Delete(id string) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *Comment) error {
	if err := r.db.Create(comment).Error; err != nil {
		return err
	}
	// Muat relasi User agar response langsung berisi data penulis komentar
	return r.db.Preload("User").First(comment, "id = ?", comment.ID).Error
}

func (r *commentRepository) FindByID(id string) (*Comment, error) {
	var comment Comment
	err := r.db.Preload("User").Where("id = ?", id).First(&comment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) FindByProjectID(projectID string) ([]Comment, error) {
	var comments []Comment
	err := r.db.Preload("User").Where("project_id = ?", projectID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *commentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Comment{}).Error
}
