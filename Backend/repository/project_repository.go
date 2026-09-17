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

	// Relasi ke galeri gambar (satu project -> banyak gambar, migration 000007)
	Gallery []ProjectGallery `gorm:"foreignKey:ProjectID" json:"gallery"`
}

// ProjectFilter menampung seluruh parameter filter list project.
// Nilai kosong berarti filter tersebut tidak aktif.
type ProjectFilter struct {
	Status    string
	TechStack string
	OwnerID   string
}

// Interface standar Clean Architecture
type ProjectRepository interface {
	Create(project *Project) error
	FindByID(id string) (*Project, error)
	FindAll(filter ProjectFilter) ([]Project, error)
	FindAllByUserID(userID string) ([]Project, error)
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
	err := r.db.Preload("User").Preload("Gallery", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC, created_at ASC, id ASC")
	}).Where("id = ?", id).First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// applyProjectFilter memasang seluruh filter ProjectFilter ke sebuah query GORM.
// Dipisah agar bisa di-reuse antara FindAll dan FindAllByUserID.
func applyProjectFilter(db *gorm.DB, filter ProjectFilter) *gorm.DB {
	query := db.Preload("User").Preload("Gallery", func(g *gorm.DB) *gorm.DB {
		return g.Order("display_order ASC, created_at ASC, id ASC")
	}).Order("created_at desc")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// tech_stack disimpan comma-separated ("React,Go") — ILIKE mencari tag
	// di posisi mana pun di dalam daftar, sehingga satu tag cocok ke banyak project.
	if filter.TechStack != "" {
		query = query.Where("tech_stack ILIKE ?", "%"+filter.TechStack+"%")
	}

	if filter.OwnerID != "" {
		query = query.Where("user_id = ?", filter.OwnerID)
	}

	return query
}

func (r *projectRepository) FindAll(filter ProjectFilter) ([]Project, error) {
	var projects []Project
	err := applyProjectFilter(r.db, filter).Find(&projects).Error
	return projects, err
}

// FindAllByUserID mengambil seluruh project milik seorang user (draft & published).
// Untuk fitur "My Projects" pada frontend.
func (r *projectRepository) FindAllByUserID(userID string) ([]Project, error) {
	var projects []Project
	err := applyProjectFilter(r.db, ProjectFilter{OwnerID: userID}).Find(&projects).Error
	return projects, err
}

func (r *projectRepository) Update(project *Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&Project{}).Error
}

// ProjectGallery menyimpan satu gambar galeri milik sebuah project.
// Relasi: satu Project -> banyak ProjectGallery (lihat migration 000007).
type ProjectGallery struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID    string    `gorm:"type:uuid;not null;index" json:"project_id"`
	ImageURL     string    `gorm:"type:varchar(255);not null" json:"image_url"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// GalleryRepository mengelola row galeri milik satu project (migration 000007).
type GalleryRepository interface {
	CreateAll(galleries []ProjectGallery) error
	DeleteAllByProjectID(projectID string) error
}

type galleryRepository struct {
	db *gorm.DB
}

func NewGalleryRepository(db *gorm.DB) GalleryRepository {
	return &galleryRepository{db: db}
}

// CreateAll menyimpan seluruh gambar galeri dalam satu operasi batch.
func (r *galleryRepository) CreateAll(galleries []ProjectGallery) error {
	return r.db.Create(&galleries).Error
}

// DeleteAllByProjectID menghapus seluruh gambar galeri milik satu project
// (dipakai untuk strategi sync: hapus semua lalu insert ulang sesuai urutan baru).
func (r *galleryRepository) DeleteAllByProjectID(projectID string) error {
	return r.db.Where("project_id = ?", projectID).Delete(&ProjectGallery{}).Error
}
