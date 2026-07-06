package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"Lullify_Backend/internal/domain/user"
)

// --- Mock repository ---

type mockRepo struct {
	users []*user.User
	err   error
}

func (m *mockRepo) Create(_ context.Context, u *user.User) error {
	if m.err != nil {
		return m.err
	}
	m.users = append(m.users, u)
	return nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	repo := &mockRepo{}
	uc := NewRegisterUseCase(repo)

	u, err := uc.Execute(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "alice@test.com",
		Password: "securepass123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("expected username alice, got %s", u.Username)
	}
	if u.PasswordHash == "securepass123" {
		t.Error("password should be hashed, not stored in clear")
	}
	if u.Role != user.RoleUser {
		t.Errorf("expected role user, got %s", u.Role)
	}
}

func TestRegister_EmptyUsername(t *testing.T) {
	repo := &mockRepo{}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), RegisterInput{
		Username: "",
		Email:    "test@test.com",
		Password: "securepass123",
	})
	if !errors.Is(err, user.ErrEmptyUsername) {
		t.Errorf("expected ErrEmptyUsername, got %v", err)
	}
}

func TestRegister_EmptyEmail(t *testing.T) {
	repo := &mockRepo{}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "",
		Password: "securepass123",
	})
	if !errors.Is(err, user.ErrEmptyEmail) {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	repo := &mockRepo{}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "notanemail",
		Password: "securepass123",
	})
	if !errors.Is(err, user.ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	repo := &mockRepo{}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "alice@test.com",
		Password: "short",
	})
	if !errors.Is(err, user.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &mockRepo{}
	uc := NewRegisterUseCase(repo)

	_, _ = uc.Execute(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "alice@test.com",
		Password: "securepass123",
	})

	_, err := uc.Execute(context.Background(), RegisterInput{
		Username: "bob",
		Email:    "alice@test.com",
		Password: "securepass456",
	})
	if !errors.Is(err, user.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegister_RepoError(t *testing.T) {
	repo := &mockRepo{err: errors.New("db down")}
	uc := NewRegisterUseCase(repo)

	_, err := uc.Execute(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "alice@test.com",
		Password: "securepass123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("securepass123"), bcrypt.DefaultCost)
	repo := &mockRepo{
		users: []*user.User{
			{
				ID:           uuid.New(),
				Username:     "alice",
				Email:        "alice@test.com",
				PasswordHash: string(hash),
				Role:         user.RoleUser,
			},
		},
	}
	uc := NewLoginUseCase(repo)

	u, err := uc.Execute(context.Background(), LoginInput{
		Email:    "alice@test.com",
		Password: "securepass123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Email != "alice@test.com" {
		t.Errorf("expected alice@test.com, got %s", u.Email)
	}
}

func TestLogin_EmptyEmail(t *testing.T) {
	repo := &mockRepo{}
	uc := NewLoginUseCase(repo)

	_, err := uc.Execute(context.Background(), LoginInput{
		Email:    "",
		Password: "securepass123",
	})
	if !errors.Is(err, user.ErrEmptyEmail) {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}

func TestLogin_EmptyPassword(t *testing.T) {
	repo := &mockRepo{}
	uc := NewLoginUseCase(repo)

	_, err := uc.Execute(context.Background(), LoginInput{
		Email:    "alice@test.com",
		Password: "",
	})
	if !errors.Is(err, user.ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockRepo{}
	uc := NewLoginUseCase(repo)

	_, err := uc.Execute(context.Background(), LoginInput{
		Email:    "nobody@test.com",
		Password: "securepass123",
	})
	if !errors.Is(err, user.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("securepass123"), bcrypt.DefaultCost)
	repo := &mockRepo{
		users: []*user.User{
			{
				ID:           uuid.New(),
				Email:        "alice@test.com",
				PasswordHash: string(hash),
			},
		},
	}
	uc := NewLoginUseCase(repo)

	_, err := uc.Execute(context.Background(), LoginInput{
		Email:    "alice@test.com",
		Password: "wrongpassword",
	})
	if !errors.Is(err, user.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
