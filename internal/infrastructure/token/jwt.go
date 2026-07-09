package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/user"
)

type JWTService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTService(secret string, access, refresh time.Duration) *JWTService {
	return &JWTService{
		secret:        []byte(secret),
		accessExpiry:  access,
		refreshExpiry: refresh,
	}
}

func (s *JWTService) GenerateTokens(u *user.User) (string, string, error) {
	access, err := s.sign(u.ID, string(u.Role), s.accessExpiry)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.sign(u.ID, string(u.Role), s.refreshExpiry)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *JWTService) ParseRefresh(refresh string) (uuid.UUID, error) {
	token, err := jwt.Parse(refresh, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid subject")
	}
	return uuid.Parse(sub)
}

func (s *JWTService) sign(userID uuid.UUID, role string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"role": role,
		"exp":  time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}
