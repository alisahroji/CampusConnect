package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Schema Project untuk database
type Project struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string    `gorm:"type:uuid;not null;index" json:"user_id"` // Relasi ke pemilik project
	Title       string    `gorm:"type:varchar(150);not null" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	TechStack   string    `gorm:"type:varchar(255)" json:"tech_stack"` // Disimpan dalam bentuk string dipisah koma (misal: "React,Go,PostgreSQL")
	RepoURL     string    `gorm:"type:varchar(255)" json:"repo_url"`
	DemoURL     string    `gorm:"type:varchar(255)" json:"demo_url"`
	ImageURL    string    `gorm:"type:varchar(255)" json:"image_url"`
	Status      string    `gorm:"type:varchar(20);default:'published'" json:"status"` // draft / published
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relasi ke tabel User (satu project dimiliki oleh satu user)
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// Interface standar Clean Architecture
type ProjectRepository interface {
	Create(project *Project) error
	FindByID(id string) (*Project, error)
	FindAll(status string) ([]Project, error)
	Update(project *Project) error
	Delete(id string) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) FindByID(id string) (*Project, error) {
	var project Project
	err := r.db.Preload("User").Where("id = ?", id).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindAll(status string) ([]Project, error) {
	var projects []Project
	query := r.db.Preload("User").Order("created_at desc")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Find(&projects).Error
	return projects, err
}

func (r *projectRepository) Update(project *Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Project{}).Error
}