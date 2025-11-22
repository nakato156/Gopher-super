package httpserver

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"goflix/api-coordinator/internal/auth"
	"goflix/api-coordinator/internal/plattform"
	"goflix/api-coordinator/internal/recommend"
	"goflix/pkg/styles"
	"goflix/pkg/types"

	"github.com/gin-gonic/gin"
)

const (
	defaultMongoRetryInterval = 15 * time.Second
)

func NewRouter(ctx context.Context, dispatchTrigger func(int) ([]types.Result, error)) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	mongoClient := connectMongoWithRetry(ctx)
	if mongoClient == nil {
		log.Print(styles.SprintfS("error", "[HTTP] No se iniciará el servidor HTTP porque no se pudo conectar a MongoDB"))
		return r
	}

	// conect with mongo
	dbName := os.Getenv("MONGO_DB_NAME")
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

	// Also expose auth routes at root so /login and /register are available
	handler.RegisterRoutes(r.Group("/"))

	// Register recommend routes (mock service initially)
	recSvc := recommend.NewService(dispatchTrigger)
	recHandler := recommend.NewHandler(recSvc)

	// Register under /api/recomend and also at root /recomend
	recHandler.RegisterRoutes(api.Group("/recomend"))
	recHandler.RegisterRoutes(r.Group("/recomend"))

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

func connectMongoWithRetry(ctx context.Context) *plattform.MongoService {
	interval := mongoRetryInterval()
	maxRetries := mongoMaxRetries()
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			log.Printf("[HTTP] Context cancelado antes de conectar a MongoDB: %v", ctx.Err())
			return nil
		default:
		}

		attempt++
		client, err := plattform.NewClient(ctx)
		if err == nil {
			if attempt > 1 {
				log.Printf("[HTTP] Conexión a MongoDB exitosa tras %d intentos", attempt)
			}
			return client
		}

		log.Printf("[HTTP] Error conectando a MongoDB (intento %d): %v", attempt, err)
		if maxRetries > 0 && attempt >= maxRetries {
			log.Printf("[HTTP] Alcanzado el máximo de intentos (%d) sin éxito", maxRetries)
			return nil
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			log.Printf("[HTTP] Context cancelado mientras se esperaba para reintentar: %v", ctx.Err())
			return nil
		}
	}
}

func mongoRetryInterval() time.Duration {
	val := strings.TrimSpace(os.Getenv("MONGO_RETRY_INTERVAL"))
	if val == "" {
		return defaultMongoRetryInterval
	}
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	log.Printf("[HTTP] Intervalo inválido para MONGO_RETRY_INTERVAL (%s), usando %s", val, defaultMongoRetryInterval)
	return defaultMongoRetryInterval
}

func mongoMaxRetries() int {
	val := strings.TrimSpace(os.Getenv("MONGO_MAX_RETRIES"))
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 0 {
		log.Printf("[HTTP] Valor inválido para MONGO_MAX_RETRIES (%s), usando ilimitado", val)
		return 0
	}
	return n
}
