package user

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"Lullify_Backend/internal/domain/user"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginUseCase struct {
	repo user.Repository
}

func NewLoginUseCase(repo user.Repository) *LoginUseCase {
	return &LoginUseCase{repo: repo}
}

func (uc *LoginUseCase) validate(input LoginInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return user.ErrEmptyEmail
	}
	if input.Password == "" {
		return user.ErrEmptyPassword
	}
	return nil
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*user.User, error) {
	if err := uc.validate(input); err != nil {
		return nil, err
	}

	u, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if u == nil {
		return nil, user.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	return u, nil
}
