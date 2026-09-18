package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ProjectBookmark menyimpan relasi user <-> project yang disimpan (bookmark).
// Schema tabel dibuat oleh migration 000008.
type ProjectBookmark struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID string    `gorm:"type:uuid;not null" json:"project_id"`
	UserID    string    `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// PostBookmark menyimpan relasi user <-> post yang disimpan (bookmark).
type PostBookmark struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    string    `gorm:"type:uuid;not null" json:"post_id"`
	UserID    string    `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectBookmarkRepository mengelola bookmark project milik user.
type ProjectBookmarkRepository interface {
	FindByUserAndProject(projectID, userID string) (*ProjectBookmark, error)
	Create(bookmark *ProjectBookmark) error
	Delete(projectID, userID string) error
	// ListByUser mengembalikan seluruh project yang di-bookmark user,
	// terbaru dulu, lengkap dengan data pemilik & galeri.
	ListByUser(userID string) ([]Project, error)
}

type projectBookmarkRepository struct {
	db *gorm.DB
}

func NewProjectBookmarkRepository(db *gorm.DB) ProjectBookmarkRepository {
	return &projectBookmarkRepository{db: db}
}

func (r *projectBookmarkRepository) FindByUserAndProject(projectID, userID string) (*ProjectBookmark, error) {
	var bookmark ProjectBookmark
	err := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&bookmark).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

func (r *projectBookmarkRepository) Create(bookmark *ProjectBookmark) error {
	return r.db.Create(bookmark).Error
}

func (r *projectBookmarkRepository) Delete(projectID, userID string) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&ProjectBookmark{}).Error
}

// ListByUser mengambil project yang di-bookmark seorang user (urut terbaru).
// Draft milik sendiri ikut tampil karena bookmark hanya bisa dibuat oleh
// pemilik bookmark itu sendiri.
func (r *projectBookmarkRepository) ListByUser(userID string) ([]Project, error) {
	var projects []Project
	err := r.db.
		Joins("JOIN project_bookmarks pb ON pb.project_id = projects.id").
		Where("pb.user_id = ?", userID).
		Preload("User").
		Preload("Gallery", func(g *gorm.DB) *gorm.DB {
			return g.Order("display_order ASC, created_at ASC, id ASC")
		}).
		Order("pb.created_at DESC").
		Find(&projects).Error
	return projects, err
}

// PostBookmarkRepository mengelola bookmark post milik user.
type PostBookmarkRepository interface {
	FindByUserAndPost(postID, userID string) (*PostBookmark, error)
	Create(bookmark *PostBookmark) error
	Delete(postID, userID string) error
	// ListByUser mengembalikan seluruh post yang di-bookmark user,
	// terbaru dulu, lengkap dengan data penulis & statistik like.
	ListByUser(userID string) ([]Post, error)
}

type postBookmarkRepository struct {
	db *gorm.DB
}

func NewPostBookmarkRepository(db *gorm.DB) PostBookmarkRepository {
	return &postBookmarkRepository{db: db}
}

func (r *postBookmarkRepository) FindByUserAndPost(postID, userID string) (*PostBookmark, error) {
	var bookmark PostBookmark
	err := r.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&bookmark).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

func (r *postBookmarkRepository) Create(bookmark *PostBookmark) error {
	return r.db.Create(bookmark).Error
}

func (r *postBookmarkRepository) Delete(postID, userID string) error {
	return r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostBookmark{}).Error
}

// ListByUser mengambil post yang di-bookmark seorang user (urut terbaru),
// lengkap dengan hydrate statistik like (like_count & liked_by_me) agar
// PostCard di halaman Bookmark menampilkan state yang benar (pola BUG-5A).
func (r *postBookmarkRepository) ListByUser(userID string) ([]Post, error) {
	var posts []Post
	err := r.db.
		Joins("JOIN post_bookmarks pbo ON pbo.post_id = posts.id").
		Where("pbo.user_id = ?", userID).
		Preload("User").
		Order("pbo.created_at DESC").
		Find(&posts).Error
	if err != nil {
		return nil, err
	}

	if err := hydratePostLikeStats(r.db, posts, userID); err != nil {
		return nil, err
	}
	return posts, nil
}
