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
