package service

import (
	"campusconnect/repository"
	"errors"
	"strings"
	"unicode/utf8"
)

type PostCommentService interface {
	AddPostComment(postID, userID, content string) (*repository.PostComment, error)
	GetPostComments(postID string) ([]repository.PostComment, error)
	DeletePostComment(commentID, userID string) error
}

type postCommentService struct {
	postCommentRepo repository.PostCommentRepository
	postRepo        repository.PostRepository
}

func NewPostCommentService(postCommentRepo repository.PostCommentRepository, postRepo repository.PostRepository) PostCommentService {
	return &postCommentService{postCommentRepo: postCommentRepo, postRepo: postRepo}
}

const maxPostCommentLength = 1000

func (s *postCommentService) AddPostComment(postID, userID, content string) (*repository.PostComment, error) {
	// 1. Pastikan post-nya ada
	if _, err := s.postRepo.FindByID(postID, ""); err != nil {
		return nil, err // Sudah berupa repository.ErrNotFound
	}

	// 2. Validasi konten komentar
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("isi komentar tidak boleh kosong")
	}
	if utf8.RuneCountInString(content) > maxPostCommentLength {
		return nil, errors.New("komentar maksimal 1000 karakter")
	}

	// 3. Simpan komentar
	comment := &repository.PostComment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}
	if err := s.postCommentRepo.Create(comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *postCommentService) GetPostComments(postID string) ([]repository.PostComment, error) {
	// Pastikan post-nya ada agar 404 konsisten
	if _, err := s.postRepo.FindByID(postID, ""); err != nil {
		return nil, err
	}
	return s.postCommentRepo.FindByPostID(postID)
}

func (s *postCommentService) DeletePostComment(commentID, userID string) error {
	// 1. Cari komentarnya dulu
	comment, err := s.postCommentRepo.FindByID(commentID)
	if err != nil {
		return err // Sudah berupa repository.ErrNotFound
	}

	// 2. Pastikan yang menghapus adalah pemiliknya
	if comment.UserID != userID {
		return repository.ErrForbidden
	}

	return s.postCommentRepo.Delete(commentID)
}
