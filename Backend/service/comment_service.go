package service

import (
	"campusconnect/repository"
	"errors"
	"strings"
)

type CommentService interface {
	AddComment(projectID, userID, content string) (*repository.Comment, error)
	GetComments(projectID string) ([]repository.Comment, error)
	DeleteComment(commentID, userID string) error
	// SetNotifier opsional (Minggu 6): membuat notification ke pemilik project.
	SetNotifier(n Notifier)
}

type commentService struct {
	commentRepo repository.CommentRepository
	projectRepo repository.ProjectRepository
	notifier    Notifier // opsional (Minggu 6)
}

func NewCommentService(commentRepo repository.CommentRepository, projectRepo repository.ProjectRepository) CommentService {
	return &commentService{commentRepo: commentRepo, projectRepo: projectRepo}
}

// SetNotifier memasang Notifier (Minggu 6) tanpa mengubah constructor existing.
func (s *commentService) SetNotifier(n Notifier) { s.notifier = n }

func (s *commentService) AddComment(projectID, userID, content string) (*repository.Comment, error) {
	// 1. Validasi konten komentar
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("isi komentar tidak boleh kosong")
	}

	// 2. Pastikan project-nya benar-benar ada
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err // Sudah berupa repository.ErrNotFound
	}

	// 3. Simpan komentar
	comment := &repository.Comment{
		ProjectID: projectID,
		UserID:    userID,
		Content:   content,
	}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}
	// Minggu 6: beri tahu pemilik project (skip bila penulis == pemilik)
	if s.notifier != nil {
		s.notifier.Notify(project.UserID, userID, repository.NotifTypeCommentProject, projectID)
	}
	return comment, nil
}

func (s *commentService) GetComments(projectID string) ([]repository.Comment, error) {
	// Pastikan project-nya ada supaya bisa mengembalikan 404 untuk project yang tidak dikenal
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		return nil, err
	}
	return s.commentRepo.FindByProjectID(projectID)
}

func (s *commentService) DeleteComment(commentID, userID string) error {
	// 1. Cari komentarnya dulu
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return err // Sudah berupa repository.ErrNotFound
	}

	// 2. Pastikan yang menghapus adalah penulis komentar
	if comment.UserID != userID {
		return repository.ErrForbidden
	}

	return s.commentRepo.Delete(commentID)
}
