package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/voxlab/voxlab-backend/internal/database"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
	rdb       *redis.Client
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
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
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		rdb:       database.GetRedis(),
	}
}

func (s *AuthService) Register(req RegisterRequest) (*LoginResponse, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("error processing password")
	}

	user := &models.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("error creating user")
	}

	return s.generateTokenResponse(user)
}

func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.generateTokenResponse(u)
}

func (s *AuthService) generateTokenResponse(u *models.User) (*LoginResponse, error) {
	expiresAt := time.Now().Add(time.Hour * 24)

	tokenID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  u.ID.String(),
		"email":    u.Email,
		"exp":      expiresAt.Unix(),
		"token_id": tokenID,
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

func (s *AuthService) Logout(tokenString string) error {
	blacklistKey := tokenBlacklistKey(tokenString)
	ttl := time.Hour * 25

	return s.rdb.Set(database.Ctx, blacklistKey, "true", ttl).Err()
}

func (s *AuthService) IsTokenBlacklisted(tokenString string) (bool, error) {
	blacklistKey := tokenBlacklistKey(tokenString)
	exists, err := s.rdb.Exists(database.Ctx, blacklistKey).Result()
	if err != nil {
		return false, fmt.Errorf("checking token blacklist: %w", err)
	}
	return exists == 1, nil
}

func (s *AuthService) GetMe(userID string) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func tokenBlacklistKey(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return "auth:blacklist:" + hex.EncodeToString(hash[:])
}
