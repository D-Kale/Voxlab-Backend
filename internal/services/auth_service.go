package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserData `json:"user"`
}

type UserData struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	XP         int       `json:"xp"`
	StreakDays int       `json:"streak_days"`
}

func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	expiresAt := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID.String(),
		"email":   u.Email,
		"exp":     expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("error generating token")
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: UserData{
			ID:         u.ID,
			Name:       u.Name,
			Email:      u.Email,
			XP:         u.XP,
			StreakDays: u.StreakDays,
		},
	}, nil
}
