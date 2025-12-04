package health

import (
	"context"
	"net/http"
	"time"

	"goflix/api-coordinator/internal/plattform"
	tcpserver "goflix/api-coordinator/internal/server/tcp"

	"github.com/gin-gonic/gin"
)

type Status struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
}

type Service interface {
	Check(ctx context.Context) Status
}

type healthService struct {
	mongoClient *plattform.MongoService
	tcpServer   *tcpserver.Server
}

func NewService(mongoClient *plattform.MongoService, tcpServer *tcpserver.Server) Service {
	return &healthService{
		mongoClient: mongoClient,
		tcpServer:   tcpServer,
	}
}

func (s *healthService) Check(ctx context.Context) Status {
	services := make(map[string]interface{})
	overallStatus := "ok"

	// 1. MongoDB Check
	mongoStatus := "ok"
	if err := s.mongoClient.Ping(ctx); err != nil {
		mongoStatus = "down"
		overallStatus = "degraded"
	}
	services["mongodb"] = map[string]string{
		"status": mongoStatus,
	}

	// 2. TCP Server & Workers Check
	// We can check if the listener is active (implicit if the app is running)
	// and count connected workers.
	workerCount := 0
	s.tcpServer.Mu.RLock()
	workerCount = len(s.tcpServer.Workers)
	s.tcpServer.Mu.RUnlock()

	services["tcp_server"] = map[string]interface{}{
		"status":       "ok", // If we are running this code, the process is up
		"worker_count": workerCount,
	}

	// 3. Frontend Check (Placeholder)
	// Since we are the backend, we can't easily know if the frontend is up unless we probe it.
	// For now, we'll assume it's "unknown" or just report that this API is ready to serve it.
	// The user asked for "frontend" status. If the frontend is a separate service (e.g. Nginx serving static files),
	// we might need to ping it. But usually healthchecks report *dependencies*.
	// Let's assume "frontend" means "is the API ready for the frontend?".
	// Or maybe we can try to make a HEAD request to the frontend URL if we knew it.
	// Given the constraints, I'll add a static "ok" or "unknown" for now, or maybe check if we are serving static files?
	// The user mentioned "front_concurrente" directory. Maybe we are serving it?
	// Looking at main.go, we don't seem to serve static files.
	// I will just report "backend_api" as "ok".
	// But the user explicitly asked for "frontend".
	// "Nos debe informar como va mongo, como va el tcp, el forntend y si hay workers o no"
	// I'll add a "frontend" key with "status": "unknown" and a note, or maybe just "ok" if the user implies the system as a whole.
	// Let's stick to what we know.
	services["frontend"] = map[string]string{
		"status": "unknown", // Cannot determine from backend
		"info":   "check frontend URL directly",
	}

	return Status{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
	}
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/health", h.HealthCheck)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	status := h.svc.Check(c.Request.Context())
	httpStatus := http.StatusOK
	if status.Status != "ok" {
		httpStatus = http.StatusServiceUnavailable
	}
	c.JSON(httpStatus, status)
}
