package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Material adalah satu dokumen/materi kuliah yang diunggah Lecturer
// (Minggu 7 Fase 3). Schema tabel dibuat oleh migration 000011; kolom
// updated_at ditambahkan migration 000012 saat fitur edit material (Day 2).
//
// File reference: file_url adalah URL publik Cloudinary (pola image_url
// existing), original_filename + file_size adalah metadata unduhan untuk UI.
// Repository CRUD disimpan di file yang sama mengikuti pola project_repository.
type Material struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UploaderID       string    `gorm:"type:uuid;not null" json:"uploader_id"`
	Title            string    `gorm:"type:varchar(150);not null" json:"title"`
	Category         string    `gorm:"type:varchar(100);not null" json:"category"`
	FileURL          string    `gorm:"type:varchar(255);not null" json:"file_url"`
	OriginalFilename string    `gorm:"type:varchar(255);not null" json:"original_filename"`
	FileSize         int64     `json:"file_size"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relasi ke tabel User (satu materi dimiliki satu uploader) — pola Project.User.
	// Dipreload manual di query; DTO handler yang menentukan field user yang tampil.
	Uploader *User `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
}

// MaterialRepository menyimpan & membaca materi kuliah.
type MaterialRepository interface {
	// Create menyimpan material baru.
	Create(material *Material) error
	// FindByID mengembalikan satu material lengkap dengan data uploader.
	// ErrNotFound bila tidak ada.
	FindByID(id string) (*Material, error)
	// List mengembalikan materi terbaru dulu, opsional difilter per kategori
	// dan per uploader. Limit dibatasi (max 100) agar query tetap bounded.
	List(category string, uploaderID string, limit int) ([]Material, error)
	// Update menyimpan perubahan material (metadata).
	Update(material *Material) error
	// Delete menghapus material. ErrNotFound bila row tidak ada.
	Delete(id string) error
}

type materialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

func (r *materialRepository) Create(material *Material) error {
	if err := r.db.Create(material).Error; err != nil {
		return err
	}
	// Preload uploader agar response create menampilkan uploader (pola sama
	// dengan FindByID); tanpa ini relasi kosong karena insert tidak memuatnya.
	return r.db.Preload("Uploader").First(material, "id = ?", material.ID).Error
}

func (r *materialRepository) FindByID(id string) (*Material, error) {
	var material Material
	err := r.db.Preload("Uploader").First(&material, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &material, nil
}

func (r *materialRepository) List(category string, uploaderID string, limit int) ([]Material, error) {
	var materials []Material
	query := r.db.Preload("Uploader").Order("created_at DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if uploaderID != "" {
		query = query.Where("uploader_id = ?", uploaderID)
	}
	// Bounded query: default 50, max 100 (pola limit notifikasi).
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if err := query.Limit(limit).Find(&materials).Error; err != nil {
		return nil, err
	}
	return materials, nil
}

func (r *materialRepository) Update(material *Material) error {
	result := r.db.Model(&Material{}).
		Where("id = ?", material.ID).
		Updates(map[string]interface{}{
			"title":    material.Title,
			"category": material.Category,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	// Refresh timestamp dari DB (GORM mengisi updated_at otomatis).
	return r.db.First(material, "id = ?", material.ID).Error
}

func (r *materialRepository) Delete(id string) error {
	result := r.db.Delete(&Material{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
