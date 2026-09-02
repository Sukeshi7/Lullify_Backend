package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/user"
)

var ErrInvalidToken = errors.New("invalid token")

const (
	typeAccess  = "access"
	typeRefresh = "refresh"
)

type Claims struct {
	UserID uuid.UUID
	Role   user.Role
}

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTService(accessSecret, refreshSecret string, access, refresh time.Duration) *JWTService {
	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpiry:  access,
		refreshExpiry: refresh,
	}
}

func (s *JWTService) GenerateTokens(u *user.User) (string, string, error) {
	access, err := s.sign(u.ID, string(u.Role), typeAccess, s.accessExpiry, s.accessSecret)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.sign(u.ID, string(u.Role), typeRefresh, s.refreshExpiry, s.refreshSecret)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *JWTService) ParseAccess(raw string) (*Claims, error) {
	return s.parse(raw, typeAccess, s.accessSecret)
}

func (s *JWTService) ParseRefresh(raw string) (*Claims, error) {
	return s.parse(raw, typeRefresh, s.refreshSecret)
}

func (s *JWTService) parse(raw, expectedType string, secret []byte) (*Claims, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if typ, _ := claims["typ"].(string); typ != expectedType {
		return nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	id, err := uuid.Parse(sub)
	if err != nil {
		return nil, ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &Claims{UserID: id, Role: user.Role(role)}, nil
}

func (s *JWTService) sign(userID uuid.UUID, role, typ string, expiry time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"role": role,
		"typ":  typ,
		"iat":  now.Unix(),
		"exp":  now.Add(expiry).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
