package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

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

func (uc *RegisterUseCase) validate(input RegisterInput) error {
	if strings.TrimSpace(input.Username) == "" {
		return user.ErrEmptyUsername
	}
	if strings.TrimSpace(input.Email) == "" {
		return user.ErrEmptyEmail
	}
	if !strings.Contains(input.Email, "@") || !strings.Contains(input.Email, ".") {
		return user.ErrInvalidEmail
	}
	if input.Password == "" {
		return user.ErrEmptyPassword
	}
	if len(input.Password) < 8 {
		return user.ErrPasswordTooShort
	}
	return nil
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*user.User, error) {
	if err := uc.validate(input); err != nil {
		return nil, err
	}

	existing, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, user.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           uuid.New(),
		Username:     strings.TrimSpace(input.Username),
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: string(hash),
		Role:         user.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return u, nil
}
