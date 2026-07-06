package user

import "errors"

var (
	ErrEmptyUsername      = errors.New("username is required")
	ErrEmptyEmail         = errors.New("email is required")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrEmptyPassword      = errors.New("password is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrEmailAlreadyExists = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
