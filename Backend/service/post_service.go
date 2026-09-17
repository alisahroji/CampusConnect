package service

import (
	"campusconnect/repository"
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	defaultPostListLimit = 10
	maxPostListLimit     = 50
)

type PostService interface {
	CreatePost(post *repository.Post) error
	GetPostByID(id string, viewerID string) (*repository.Post, error)
	// UpdatePost memperbarui konten post milik userID. Ownership divalidasi
	// di service agar terproteksi untuk semua pemanggil.
	UpdatePost(postID string, userID string, updates *repository.Post) (*repository.Post, error)
	// GetAllPosts mengambil feed explore global dengan cursor pagination.
	// viewerID (boleh kosong = anonymous) dipakai untuk mengisi liked_by_me.
	// Mengembalikan nextCursor ("") jika sudah halaman terakhir.
	GetAllPosts(cursor string, limit int, viewerID string) ([]repository.Post, string, error)
	DeletePost(postID string, userID string) error
}

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

const maxPostLength = 2000

func (s *postService) CreatePost(post *repository.Post) error {
	// Validasi dasar: konten wajib ada setelah trim spasi
	post.Content = strings.TrimSpace(post.Content)
	if post.Content == "" {
		return errors.New("konten post tidak boleh kosong")
	}

	// Batasi panjang konten agar feed tetap sehat
	if utf8.RuneCountInString(post.Content) > maxPostLength {
		return errors.New("konten post maksimal 2000 karakter")
	}

	// image_url boleh kosong; kalau ada, trim sekalian
	post.ImageURL = strings.TrimSpace(post.ImageURL)

	return s.repo.Create(post)
}

func (s *postService) GetPostByID(id string, viewerID string) (*repository.Post, error) {
	return s.repo.FindByID(id, viewerID)
}

// UpdatePost memperbarui konten post milik userID. Mengikuti pola project:
// validasi konten lalu ownership check (403 untuk non-owner, 404 untuk post
// yang tidak ada). Mengembalikan post terbaru (beserta statistik like).
func (s *postService) UpdatePost(postID string, userID string, updates *repository.Post) (*repository.Post, error) {
	// 1. Cari post-nya dulu (repository.ErrNotFound bila tidak ada)
	post, err := s.repo.FindByID(postID, userID)
	if err != nil {
		return nil, err
	}

	// 2. Ownership check
	if post.UserID != userID {
		return nil, repository.ErrForbidden
	}

	// 3. Validasi konten yang sama dengan create
	newContent := strings.TrimSpace(updates.Content)
	if newContent == "" {
		return nil, errors.New("konten post tidak boleh kosong")
	}
	if utf8.RuneCountInString(newContent) > maxPostLength {
		return nil, errors.New("konten post maksimal 2000 karakter")
	}

	// 4. Terapkan perubahan & simpan
	post.Content = newContent
	post.ImageURL = strings.TrimSpace(updates.ImageURL)
	if err := s.repo.Update(post); err != nil {
		return nil, err
	}

	// 5. Kembalikan data terbaru (Preload User + statistik like)
	return s.repo.FindByID(postID, userID)
}

func (s *postService) GetAllPosts(cursor string, limit int, viewerID string) ([]repository.Post, string, error) {
	if limit <= 0 {
		limit = defaultPostListLimit
	}
	if limit > maxPostListLimit {
		limit = maxPostListLimit
	}
	return s.repo.FindAll(cursor, limit, viewerID)
}

func (s *postService) DeletePost(postID string, userID string) error {
	// 1. Cari post-nya dulu
	post, err := s.repo.FindByID(postID, "")
	if err != nil {
		return err // Sudah berupa repository.ErrNotFound
	}

	// 2. Pastikan yang menghapus adalah pemiliknya
	if post.UserID != userID {
		return repository.ErrForbidden
	}

	return s.repo.Delete(postID)
}
