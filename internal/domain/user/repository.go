package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)

	FindAll(ctx context.Context) ([]*User, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	CountAll(ctx context.Context) (int, error)
	CountByRole(ctx context.Context, role Role) (int, error)
}
