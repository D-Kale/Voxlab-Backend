package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/voxlab/voxlab-backend/internal/domain/user"
)

type Service struct {
	userRepo  *user.Repository
	jwtSecret string
}

func NewService(userRepo *user.Repository, jwtSecret string) *Service {
	return &Service{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *Service) Login(email, password string) (*LoginResponse, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	expiresAt := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID.String(),
		"email":   u.Email,
		"exp":     expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("error generando token")
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
