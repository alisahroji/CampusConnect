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
	GetPostByID(id string) (*repository.Post, error)
	// GetAllPosts mengambil feed explore global dengan cursor pagination.
	// Mengembalikan nextCursor ("") jika sudah halaman terakhir.
	GetAllPosts(cursor string, limit int) ([]repository.Post, string, error)
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

func (s *postService) GetPostByID(id string) (*repository.Post, error) {
	return s.repo.FindByID(id)
}

func (s *postService) GetAllPosts(cursor string, limit int) ([]repository.Post, string, error) {
	if limit <= 0 {
		limit = defaultPostListLimit
	}
	if limit > maxPostListLimit {
		limit = maxPostListLimit
	}
	return s.repo.FindAll(cursor, limit)
}

func (s *postService) DeletePost(postID string, userID string) error {
	// 1. Cari post-nya dulu
	post, err := s.repo.FindByID(postID)
	if err != nil {
		return err // Sudah berupa repository.ErrNotFound
	}

	// 2. Pastikan yang menghapus adalah pemiliknya
	if post.UserID != userID {
		return repository.ErrForbidden
	}

	return s.repo.Delete(postID)
}
