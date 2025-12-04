package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// service implementa Service.
type service struct {
	repo   Repository
	tokens TokenManager
	now    func() time.Time
}

// NewService construye el servicio de autenticación.
func NewService(repo Repository, tokens TokenManager) Service {
	return &service{
		repo:   repo,
		tokens: tokens,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (s *service) Register(ctx context.Context, email, password string) (string, string, error) {
	fmt.Println("Registering user with email:", email)
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return "", "", ErrUserAlreadyExists
	} else if err != ErrUserNotFound {
		return "", "", err
	}

	fmt.Println("No existing user found with email:", email)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	u := &User{
		Email:        email,
		PasswordHash: string(hash),
		UserID:       uuid.New().String(),
	}

	if err := s.repo.CreateUser(ctx, u); err != nil {
		return "", "", err
	}

	token, err := s.tokens.GenerateToken(u.ID.Hex(), u.UserID)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("User registered with ID: %s\n", u.ID.Hex())
	return u.ID.Hex(), token, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, error) {
	fmt.Printf("searching %s", email)
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == ErrUserNotFound {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	token, err := s.tokens.GenerateToken(u.ID.Hex(), u.UserID)
	if err != nil {
		return "", "", err
	}

	return u.ID.Hex(), token, nil
}
