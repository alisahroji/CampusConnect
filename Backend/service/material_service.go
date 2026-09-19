package service

import (
	"campusconnect/repository"
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// FileUploader abstraction agar service tidak bergantung langsung ke Cloudinary
// (dan bisa di-mock di unit test). Implementasi nyata ada di handler layer
// (cloudinaryUploader) karena pembacaan kredensial env mengikuti pola upload
// existing (user_handler / image_handler).
type FileUploader interface {
	// Upload mengunggah konten file ke storage dan mengembalikan URL publiknya.
	Upload(filename string, size int64, content io.Reader) (url string, err error)
}

// Konstanta validasi material (keputusan implementasi Day 2, bukan angka dari
// roadmap — dokumentasi lengkap di laporan Day 2).
const (
	MaxMaterialSize = 20 * 1024 * 1024 // 20 MB: cukup untuk slide/dokumen kuliah
	MaxTitleLen     = 150              // = lebar kolom materials.title (000011)
	MaxCategoryLen  = 100              // = lebar kolom materials.category (000011)
)

// allowedMaterialExtensions: allowlist dokumen/slide/arsip gambar yang relevan
// untuk materi kuliah. Executable (exe/bat/sh/apk/dll) sengaja TIDAK diterima.
var allowedMaterialExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".ppt":  true,
	".pptx": true,
	".xls":  true,
	".xlsx": true,
	".txt":  true,
	".zip":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// MaterialService berisi aturan bisnis material:
// upload khusus Lecturer, ownership untuk update/delete.
type MaterialService interface {
	// CreateMaterial: validasi + upload file + simpan metadata.
	CreateMaterial(uploaderID string, role string, title string, category string, filename string, size int64, content io.Reader) (*repository.Material, error)
	// GetMaterial / ListMaterials: authenticated read (any role).
	GetMaterial(id string) (*repository.Material, error)
	ListMaterials(category string, uploaderID string, limit int) ([]repository.Material, error)
	// UpdateMaterial: hanya uploader (owner) boleh mengubah metadata.
	UpdateMaterial(materialID string, userID string, title string, category string) (*repository.Material, error)
	// DeleteMaterial: hanya uploader (owner) boleh menghapus.
	DeleteMaterial(materialID string, userID string) error
}

type materialService struct {
	repo     repository.MaterialRepository
	uploader FileUploader
}

func NewMaterialService(repo repository.MaterialRepository, uploader FileUploader) MaterialService {
	return &materialService{repo: repo, uploader: uploader}
}

// validateMaterialInput membersihkan & memvalidasi title/category.
func validateMaterialInput(title string, category string) (string, string, error) {
	title = strings.TrimSpace(title)
	category = strings.TrimSpace(category)
	if title == "" {
		return "", "", errors.New("judul materi wajib diisi")
	}
	if len(title) > MaxTitleLen {
		return "", "", errors.New("judul materi maksimal 150 karakter")
	}
	if category == "" {
		return "", "", errors.New("kategori/mata kuliah wajib diisi")
	}
	if len(category) > MaxCategoryLen {
		return "", "", errors.New("kategori/mata kuliah maksimal 100 karakter")
	}
	return title, category, nil
}

// validateFile memvalidasi ekstensi & ukuran file sebelum dikirim ke storage.
// Content-Type dari client TIDAK dipercaya (mudah dipalsukan); ekstensi nama
// file hanya metadata rapi — keputusan accept/reject berbasis allowlist.
func validateFile(filename string, size int64) (string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return "", errors.New("file materi wajib dipilih")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || !allowedMaterialExtensions[ext] {
		return "", errors.New("format file tidak didukung (boleh: PDF, DOC/DOCX, PPT/PPTX, XLS/XLSX, TXT, ZIP, JPG, PNG)")
	}
	if size <= 0 {
		return "", errors.New("file tidak boleh kosong")
	}
	if size > MaxMaterialSize {
		return "", errors.New("ukuran file maksimal 20 MB")
	}
	return ext, nil
}

func (s *materialService) CreateMaterial(uploaderID string, role string, title string, category string, filename string, size int64, content io.Reader) (*repository.Material, error) {
	// Roadmap Day 3: form upload khusus Lecturer. Role berasal dari currentUser
	// hasil RequireAuth (DB), bukan dari payload client.
	if role != "Lecturer" {
		return nil, repository.ErrForbidden
	}

	title, category, err := validateMaterialInput(title, category)
	if err != nil {
		return nil, err
	}
	if _, err := validateFile(filename, size); err != nil {
		return nil, err
	}

	// Upload file ke storage dulu (jangan simpan metadata sebelum file ada).
	fileURL, err := s.uploader.Upload(filename, size, content)
	if err != nil {
		// Jangan bocorkan detail provider (kredensial/internal) ke client.
		return nil, errors.New("gagal mengunggah file ke penyimpanan cloud")
	}

	material := &repository.Material{
		UploaderID:       uploaderID,
		Title:            title,
		Category:         category,
		FileURL:          fileURL,
		OriginalFilename: strings.TrimSpace(filename),
		FileSize:         size,
	}
	if err := s.repo.Create(material); err != nil {
		// Catatan: file sudah masuk storage tapi DB gagal -> potensi orphan file.
		// Cloudinary deletion belum ada di architecture existing, jadi tradeoff
		// ini didokumentasikan (bukan disembunyikan) di laporan Day 2.
		return nil, errors.New("gagal menyimpan data materi")
	}
	return material, nil
}

func (s *materialService) GetMaterial(id string) (*repository.Material, error) {
	return s.repo.FindByID(id)
}

func (s *materialService) ListMaterials(category string, uploaderID string, limit int) ([]repository.Material, error) {
	return s.repo.List(strings.TrimSpace(category), uploaderID, limit)
}

func (s *materialService) UpdateMaterial(materialID string, userID string, title string, category string) (*repository.Material, error) {
	title, category, err := validateMaterialInput(title, category)
	if err != nil {
		return nil, err
	}

	material, err := s.repo.FindByID(materialID)
	if err != nil {
		return nil, err // ErrNotFound diteruskan
	}
	// Ownership: hanya uploader asli boleh mengubah.
	if material.UploaderID != userID {
		return nil, repository.ErrForbidden
	}

	material.Title = title
	material.Category = category
	if err := s.repo.Update(material); err != nil {
		return nil, err
	}
	return material, nil
}

func (s *materialService) DeleteMaterial(materialID string, userID string) error {
	material, err := s.repo.FindByID(materialID)
	if err != nil {
		return err
	}
	if material.UploaderID != userID {
		return repository.ErrForbidden
	}
	return s.repo.Delete(materialID)
}
