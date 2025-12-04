package auth

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User representa al usuario en la base de datos.
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       string        `bson:"userId" json:"userId"`
	Email        string        `bson:"email" json:"email"`
	PasswordHash string        `bson:"password" json:"-"`
	// es una map {string: float64}
	Ratings map[string]float64 `bson:"ratings,omitempty" json:"ratings,omitempty"`
}

// Errores de dominio de auth.
var (
	ErrUserAlreadyExists  = errors.New("auth: user already exists")
	ErrUserNotFound       = errors.New("auth: user not found")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)

// Repository define las operaciones necesarias contra la persistencia.
type Repository interface {
	CreateUser(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// Service define la lógica de negocio expuesta a los handlers.
type Service interface {
	Register(ctx context.Context, email, password string) (userID, token string, err error)
	Login(ctx context.Context, email, password string) (userID, token string, err error)
}

// TokenManager abstrae la generación de tokens (JWT o lo que quieras).
type TokenManager interface {
	GenerateToken(userID string, appUserID string) (string, error)
	ValidateToken(token string) (string, error)
}
