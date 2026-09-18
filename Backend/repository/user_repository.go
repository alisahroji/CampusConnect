package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Definisikan struct User agar repo bisa mengaksesnya
type User struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email       string    `gorm:"unique;not null" json:"email"`
	Name        string    `gorm:"not null" json:"name"`
	PictureURL  string    `json:"picture_url"`
	Role        string    `gorm:"default:'Student'" json:"role"`
	// Minggu 6 Hari 4: status blokir akun (Admin User Management).
	Banned      bool      `gorm:"not null;default:false" json:"banned"`
	Bio         string    `json:"bio"`
	Skills      string    `json:"skills"`
	GithubURL   string    `json:"github_url"`
	LinkedinURL string    `json:"linkedin_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserRepository interface {
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	// Minggu 6 Hari 4: dukungan Admin User Management.
	List(limit int) ([]User, error)
	UpdateBanned(id string, banned bool) error
	UpdateRole(id string, role string) error
	CountByRole(role string) (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// List mengembalikan daftar user untuk admin (terbaru dulu), dibatasi limit.
func (r *userRepository) List(limit int) ([]User, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var users []User
	err := r.db.Order("created_at DESC").Limit(limit).Find(&users).Error
	return users, err
}

// UpdateBanned mengubah hanya kolom banned (targeted update, tanpa menyentuh
// field lain — lebih aman daripada Save seluruh struct).
func (r *userRepository) UpdateBanned(id string, banned bool) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("banned", banned)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateRole mengubah hanya kolom role.
func (r *userRepository) UpdateRole(id string, role string) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountByRole menghitung jumlah user dengan role tertentu (dipakai guard
// "minimal satu Admin" di service).
func (r *userRepository) CountByRole(role string) (int64, error) {
	var total int64
	err := r.db.Model(&User{}).Where("role = ?", role).Count(&total).Error
	return total, err
}
