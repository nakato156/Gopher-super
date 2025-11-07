package httpserver

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"goflix/api-coordinator/database"
	"goflix/api-coordinator/internal/httpserver/routes/auth"
)

// Server orquesta las dependencias del servidor HTTP, siguiendo un enfoque de capas.
type Server struct {
	engine *gin.Engine
	db     *database.MongoService
}

// New construye una instancia de Server inicializando dependencias y rutas.
func New(ctx context.Context) (*Server, error) {
	service, err := database.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("httpserver: conectar a mongo: %w", err)
	}

	authRepo := auth.NewMongoRepository(service, "sample_mflix", "users")
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	router := gin.Default()
	api := router.Group("/api/v1")
	authHandler.RegisterRoutes(api.Group("/auth"))

	return &Server{
		engine: router,
		db:     service,
	}, nil
}

// Engine expone el *gin.Engine subyacente (útil para pruebas).
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Run inicia el servidor HTTP en la dirección indicada.
func (s *Server) Run(addr string) error {
	if addr == "" {
		addr = ":8080"
	}
	return s.engine.Run(addr)
}

// Shutdown libera la conexión a MongoDB.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	return s.db.Disconnect(ctx)
}
