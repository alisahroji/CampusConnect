package repository

import (
	"strings"

	"gorm.io/gorm"
)

// escapeLikePattern menormalkan input pencarian agar diperlakukan sebagai
// teks biasa, bukan pola LIKE: karakter %, _ dan \ di-escape sehingga query
// seperti "q=%" tidak lagi match semua baris. Dipakai bersama ESCAPE '\\'.
// Dicari per rune agar escape-nya benar untuk karakter multi-byte.
func escapeLikePattern(input string) string {
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if r == '\\' || r == '%' || r == '_' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// SearchRepository menangani pencarian dasar lintas entitas (Minggu 5 Hari 4).
// Dibuat sebagai interface terpisah agar tidak mengubah UserRepository yang
// sudah dikunci oleh unit test mock.
type SearchRepository interface {
	SearchUsers(query string, limit int) ([]User, error)
	SearchProjects(query string, limit int) ([]Project, error)
}

type searchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) SearchRepository {
	return &searchRepository{db: db}
}

// SearchUsers mencari user berdasarkan nama (case-insensitive).
// Karakter wildcard LIKE dari input user di-escape sehingga diperlakukan
// sebagai teks biasa.
func (r *searchRepository) SearchUsers(query string, limit int) ([]User, error) {
	var users []User
	pattern := "%" + escapeLikePattern(query) + "%"
	// ESCAPE '\\' di source Go = ESCAPE '\\' di SQL (satu karakter backslash)
	err := r.db.
		Where("name ILIKE ? ESCAPE '\\'", pattern).
		Order("name asc").
		Limit(limit).
		Find(&users).Error
	return users, err
}

// SearchProjects mencari project berdasarkan judul atau tech_stack.
// Wildcard LIKE dari input user di-escape (lihat escapeLikePattern).
func (r *searchRepository) SearchProjects(query string, limit int) ([]Project, error) {
	var projects []Project
	pattern := "%" + escapeLikePattern(query) + "%"
	err := r.db.
		Preload("User").
		Where("title ILIKE ? ESCAPE '\\' OR tech_stack ILIKE ? ESCAPE '\\'", pattern, pattern).
		Order("created_at desc").
		Limit(limit).
		Find(&projects).Error
	return projects, err
}
