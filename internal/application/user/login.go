package user

import (
	"context"
	"errors"

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

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*user.User, error) {
	u, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return u, nil
}
