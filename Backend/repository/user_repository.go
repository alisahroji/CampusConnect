package repository

import (
	"gorm.io/gorm"
)

// Definisikan struct User agar repo bisa mengaksesnya
type User struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email       string `gorm:"unique;not null"`
	Name        string `gorm:"not null"`
	PictureURL  string
	Role        string `gorm:"default:'Student'"`
	Bio         string // Kolom baru
	Skills      string // Kolom baru
	GithubURL   string // Kolom baru
	LinkedinURL string // Kolom baru
}

type UserRepository interface {
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	Create(user *User) error
	Update(user *User) error
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
