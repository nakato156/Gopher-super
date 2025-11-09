package plattform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	// ErrMissingMongoURI indicates that the expected environment variable is not set.
	ErrMissingMongoURI = errors.New("database: missing MONGODB_URI environment variable")
)

// NewClient establishes a MongoDB client and returns a MongoService.
// The caller owns the returned service and must call service.client.Disconnect when done.
func NewClient(ctx context.Context) (*MongoService, error) {
	uri := strings.TrimSpace(os.Getenv("MONGODB_URI"))
	if uri == "" {
		return nil, fmt.Errorf("%w", ErrMissingMongoURI)
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opt := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opt)
	if err != nil {
		return nil, fmt.Errorf("database: connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return NewMongoService(client), nil
}

// GetCollection returns a handle to the requested collection using the provided client.
// This helper exists to make it explicit where collections are obtained from.
func GetCollection(client *mongo.Client, dbName, collName string) *mongo.Collection {
	return client.Database(dbName).Collection(collName)
}

type MongoService struct {
	client *mongo.Client
}

// NewMongoService creates a new MongoService instance with the provided MongoDB client.
func NewMongoService(client *mongo.Client) *MongoService {
	return &MongoService{client: client}
}

// InsertOne inserts a single document into the specified collection.
func (s *MongoService) InsertOne(ctx context.Context, dbName, collName string, doc interface{}) (*mongo.InsertOneResult, error) {
	coll := GetCollection(s.client, dbName, collName)
	return coll.InsertOne(ctx, doc)
}

// FindOne finds a single document in the specified collection.
func (s *MongoService) FindOne(ctx context.Context, dbName, collName string, filter interface{}) *mongo.SingleResult {
	coll := GetCollection(s.client, dbName, collName)
	return coll.FindOne(ctx, filter)
}

// Find finds multiple documents in the specified collection.
func (s *MongoService) Find(ctx context.Context, dbName, collName string, filter interface{}) (*mongo.Cursor, error) {
	coll := GetCollection(s.client, dbName, collName)
	return coll.Find(ctx, filter)
}

// UpdateOne updates a single document in the specified collection.
func (s *MongoService) UpdateOne(ctx context.Context, dbName, collName string, filter, update interface{}) (*mongo.UpdateResult, error) {
	coll := GetCollection(s.client, dbName, collName)
	return coll.UpdateOne(ctx, filter, update)
}

// GetCollection returns a handle to the requested collection.
func (s *MongoService) GetCollection(dbName, collName string) *mongo.Collection {
	return s.client.Database(dbName).Collection(collName)
}
