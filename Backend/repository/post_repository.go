package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema Post untuk database (Minggu 5 - Feed & Social Graph)
type Post struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"` // Relasi ke penulis post
	Content   string    `gorm:"type:text;not null" json:"content"`
	ImageURL  string    `gorm:"type:varchar(255)" json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relasi ke tabel User (satu post dimiliki oleh satu user)
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// Interface standar Clean Architecture
type PostRepository interface {
	Create(post *Post) error
	FindByID(id string) (*Post, error)
	// FindAll mengambil post global (explore) dengan cursor pagination.
	// Mengembalikan nextCursor kosong ("") jika sudah halaman terakhir.
	FindAll(cursor string, limit int) ([]Post, string, error)
	// FindByUserIDs mengembalikan post dari sekumpulan user tertentu
	// (dipakai untuk feed berbasis following) dengan cursor pagination.
	FindByUserIDs(userIDs []string, cursor string, limit int) ([]Post, string, error)
	Delete(id string) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *Post) error {
	if err := r.db.Create(post).Error; err != nil {
		return err
	}
	// Muat relasi User agar response langsung berisi data penulis post
	return r.db.Preload("User").First(post, "id = ?", post.ID).Error
}

func (r *postRepository) FindByID(id string) (*Post, error) {
	var post Post
	err := r.db.Preload("User").Where("id = ?", id).First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// paginatePosts menjalankan query keyset pagination: ambil limit+1 baris
// untuk mendeteksi halaman berikutnya, lalu susun nextCursor dari elemen
// terakhir halaman ini. Perbandingan (created_at, id) membuat cursor aman
// dari duplikat/loncatan ketika dua post memiliki timestamp sama persis.
func (r *postRepository) paginatePosts(query *gorm.DB, cursor string, limit int) ([]Post, string, error) {
	if cursor != "" {
		cursorTime, cursorID, err := DecodePostCursor(cursor)
		if err != nil {
			return nil, "", err
		}
		// Row-value comparison: ambil yang LEBIH LAMA dari cursor
		query = query.Where("(created_at, id) < (?, ?)", cursorTime, cursorID)
	}

	var posts []Post
	err := query.Preload("User").
		Order("created_at desc, id desc").
		Limit(limit + 1). // +1 sebagai "probe" deteksi halaman berikutnya
		Find(&posts).Error
	if err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(posts) > limit {
		// Ada halaman berikutnya: buang probe, susun cursor dari post terakhir
		posts = posts[:limit]
		last := posts[len(posts)-1]
		nextCursor = EncodePostCursor(last.CreatedAt, last.ID)
	}

	return posts, nextCursor, nil
}

func (r *postRepository) FindAll(cursor string, limit int) ([]Post, string, error) {
	return r.paginatePosts(r.db.Model(&Post{}), cursor, limit)
}

// FindByUserIDs mengambil post hanya dari user-user dalam daftar (feed following).
// Jika daftar kosong, kembalikan halaman kosong (bukan seluruh post).
func (r *postRepository) FindByUserIDs(userIDs []string, cursor string, limit int) ([]Post, string, error) {
	if len(userIDs) == 0 {
		return []Post{}, "", nil
	}
	query := r.db.Model(&Post{}).Where("user_id IN ?", userIDs)
	return r.paginatePosts(query, cursor, limit)
}

func (r *postRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Post{}).Error
}
