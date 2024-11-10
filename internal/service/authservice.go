package service

import (
	"errors"
	"fmt"
	"github.com/apolunin/slotgame/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

const (
	claimLogin = "login"
	claimExp   = "exp"
)

type AuthService struct {
	jwtSecret []byte
}

func NewAuthService(jwtSecret []byte) *AuthService {
	return &AuthService{
		jwtSecret: jwtSecret,
	}
}

func (as *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

func (as *AuthService) IsValidPassword(hash, password string) bool {
	switch err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); {
	case err == nil:
		return true
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false
	default:
		slog.With(
			logger.FieldError, err,
			logger.FieldHash, hash,
		).Error("failed to validate hash")

		return false
	}
}

func (as *AuthService) CreateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			claimLogin: login,
			claimExp:   time.Now().Add(time.Hour * 24).Unix(),
		},
	)

	tokenString, err := token.SignedString(as.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (as *AuthService) GetLoginFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return as.jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse jwt token: %w", err)
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims format")
	}

	login, ok := claims[claimLogin].(string)
	if !ok {
		return "", errors.New("login claim is missing or has invalid format")
	}

	return login, nil
}
