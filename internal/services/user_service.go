package services

import (
	"time"

	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type UserService struct {
	repo *repositories.UserRepository
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type AdminUserData struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	XP         int       `json:"xp"`
	StreakDays int       `json:"streak_days"`
	Lives      int       `json:"lives"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Role  *string `json:"role,omitempty"`
}

func (s *UserService) ListUsers() ([]AdminUserData, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	result := make([]AdminUserData, len(users))
	for i, u := range users {
		result[i] = toAdminData(u)
	}
	return result, nil
}

func (s *UserService) GetUser(id string) (*AdminUserData, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	data := toAdminData(*u)
	return &data, nil
}

func (s *UserService) UpdateUser(id string, req UpdateUserRequest) (*AdminUserData, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Email != nil {
		u.Email = *req.Email
	}
	if req.Role != nil {
		u.Role = *req.Role
	}

	if err := s.repo.Update(u); err != nil {
		return nil, err
	}

	data := toAdminData(*u)
	return &data, nil
}

func (s *UserService) DeleteUser(id string) error {
	return s.repo.Delete(id)
}

func toAdminData(u models.User) AdminUserData {
	return AdminUserData{
		ID:         u.ID.String(),
		Name:       u.Name,
		Email:      u.Email,
		Role:       u.Role,
		XP:         u.XP,
		StreakDays: u.StreakDays,
		Lives:      u.Lives,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}
