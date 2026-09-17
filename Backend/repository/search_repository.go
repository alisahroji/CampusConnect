package repository

import (
	"gorm.io/gorm"
)

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
func (r *searchRepository) SearchUsers(query string, limit int) ([]User, error) {
	var users []User
	pattern := "%" + query + "%"
	err := r.db.
		Where("name ILIKE ?", pattern).
		Order("name asc").
		Limit(limit).
		Find(&users).Error
	return users, err
}

// SearchProjects mencari project berdasarkan judul atau tech_stack.
func (r *searchRepository) SearchProjects(query string, limit int) ([]Project, error) {
	var projects []Project
	pattern := "%" + query + "%"
	err := r.db.
		Preload("User").
		Where("title ILIKE ? OR tech_stack ILIKE ?", pattern, pattern).
		Order("created_at desc").
		Limit(limit).
		Find(&projects).Error
	return projects, err
}
