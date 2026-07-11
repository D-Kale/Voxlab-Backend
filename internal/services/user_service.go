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
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name       string    `json:"name" example:"John Doe"`
	Email      string    `json:"email" example:"john@example.com"`
	Role       string    `json:"role" example:"user"`
	XP         int       `json:"xp" example:"850"`
	StreakDays int       `json:"streak_days" example:"3"`
	Lives      int       `json:"lives" example:"5"`
	CreatedAt  time.Time `json:"created_at" example:"2025-06-01T12:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2026-06-18T12:00:00Z"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" example:"Jane Doe"`
	Email *string `json:"email,omitempty" example:"jane@example.com"`
	Role  *string `json:"role,omitempty" example:"admin"`
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

type LeaderboardUser struct {
	Name string `json:"name" example:"John Doe"`
	XP   int    `json:"xp" example:"850"`
}

type LeaderboardData struct {
	MyRank  int               `json:"my_rank" example:"3"`
	MyXP    int               `json:"my_xp" example:"850"`
	TopUsers []LeaderboardUser `json:"top_users"`
}

func (s *UserService) GetLeaderboard(userID string, limit int) (*LeaderboardData, error) {
	users, err := s.repo.FindLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	topUsers := make([]LeaderboardUser, len(users))
	for i, u := range users {
		topUsers[i] = LeaderboardUser{
			Name: u.Name,
			XP:   u.XP,
		}
	}

	rank, err := s.repo.CountByXPGreaterThan(user.XP)
	if err != nil {
		rank = 0
	}

	return &LeaderboardData{
		MyRank:   int(rank) + 1,
		MyXP:     user.XP,
		TopUsers: topUsers,
	}, nil
}
