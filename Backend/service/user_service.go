package service

import (
	"campusconnect/repository" // Sesuaikan dengan nama modul di go.mod kamu
)

type UserService interface {
	GetUserProfile(userID string) (*repository.User, error)
	UpdateProfile(user *repository.User) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{userRepo: repo}
}

func (s *userService) GetUserProfile(userID string) (*repository.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *userService) UpdateProfile(user *repository.User) error {
	return s.userRepo.Update(user)
}
