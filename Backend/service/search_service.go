package service

import (
	"campusconnect/repository"
	"errors"
	"strings"
)

const (
	defaultSearchLimit = 10
	maxSearchLimit     = 25
)

// SearchResult adalah DTO gabungan untuk endpoint pencarian.
type SearchResult struct {
	Users    []repository.User    `json:"users"`
	Projects []repository.Project `json:"projects"`
}

type SearchService interface {
	Search(query string, limit int) (*SearchResult, error)
}

type searchService struct {
	repo repository.SearchRepository
}

func NewSearchService(repo repository.SearchRepository) SearchService {
	return &searchService{repo: repo}
}

// Search mencari user (by name) dan project (by title/tech_stack)
// dalam satu panggilan, dengan limit per entitas.
func (s *searchService) Search(query string, limit int) (*SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query pencarian tidak boleh kosong")
	}

	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	users, err := s.repo.SearchUsers(query, limit)
	if err != nil {
		return nil, err
	}

	projects, err := s.repo.SearchProjects(query, limit)
	if err != nil {
		return nil, err
	}

	// Pastikan selalu array, bukan null, agar enak dikonsumsi frontend
	if users == nil {
		users = []repository.User{}
	}
	if projects == nil {
		projects = []repository.Project{}
	}

	return &SearchResult{Users: users, Projects: projects}, nil
}
