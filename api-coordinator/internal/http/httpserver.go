package httpserver

import (
	"context"
	"log"
	"os"

	"goflix/api-coordinator/internal/auth"
	"goflix/api-coordinator/internal/plattform"
	"goflix/pkg/styles"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// platform services
	ctx := context.Background()
	mongoClient, err := plattform.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// conect with mongo
	dbName := "goflixx"
	usersColl := mongoClient.GetCollection(dbName, "users")
	repo := auth.NewMongoRepository(usersColl)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key" // default
	}

	tokenManager := auth.NewJWTTokenManager(secret)
	svc := auth.NewService(repo, tokenManager)
	handler := auth.NewHandler(svc)

	// Register auth routes
	api := r.Group("/api")
	authGroup := api.Group("/auth")
	handler.RegisterRoutes(authGroup)

	// Escuchar en todas las interfaces del contenedor
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":80"
	}

	log.Print(styles.SprintfS("info", "[HTTP] Escuchando en %s", addr))
	if err := r.Run(addr); err != nil {
		log.Fatal(styles.SprintfS("error", "[HTTP] Error: %v", err))
	}

	return r
}
