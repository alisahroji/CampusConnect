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
	// SetNotifier opsional (Minggu 6): membuat notification ke penulis post.
	SetNotifier(n Notifier)
}

type postCommentService struct {
	postCommentRepo repository.PostCommentRepository
	postRepo        repository.PostRepository
	notifier        Notifier // opsional (Minggu 6)
}

func NewPostCommentService(postCommentRepo repository.PostCommentRepository, postRepo repository.PostRepository) PostCommentService {
	return &postCommentService{postCommentRepo: postCommentRepo, postRepo: postRepo}
}

// SetNotifier memasang Notifier (Minggu 6) tanpa mengubah constructor existing.
func (s *postCommentService) SetNotifier(n Notifier) { s.notifier = n }

const maxPostCommentLength = 1000

func (s *postCommentService) AddPostComment(postID, userID, content string) (*repository.PostComment, error) {
	// 1. Pastikan post-nya ada
	post, err := s.postRepo.FindByID(postID, "")
	if err != nil {
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
	// Minggu 6: beri tahu penulis post (skip bila penulis komentar == penulis post)
	if s.notifier != nil {
		s.notifier.Notify(post.UserID, userID, repository.NotifTypeCommentPost, postID)
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
