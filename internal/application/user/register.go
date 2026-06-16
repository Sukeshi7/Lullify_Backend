package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"Lullify_Backend/internal/domain/user"
)

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type RegisterUseCase struct {
	repo user.Repository
}

func NewRegisterUseCase(repo user.Repository) *RegisterUseCase {
	return &RegisterUseCase{repo: repo}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*user.User, error) {
	existing, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: input.Password,
		Role:         user.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
