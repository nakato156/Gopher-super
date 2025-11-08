package auth

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mongoRepository struct {
	coll *mongo.Collection
}

// NewMongoRepository crea un repositorio de usuarios basado en MongoDB.
func NewMongoRepository(coll *mongo.Collection) Repository {
	return &mongoRepository{coll: coll}
}

func (r *mongoRepository) CreateUser(ctx context.Context, u *User) error {
	u.Email = normalizeEmail(u.Email)

	_, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *mongoRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	email = normalizeEmail(email)

	var u User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func normalizeEmail(e string) string {
	return strings.TrimSpace(strings.ToLower(e))
}
