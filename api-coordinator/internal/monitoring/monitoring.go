package monitoring

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"goflix/api-coordinator/internal/plattform"
	tcpserver "goflix/api-coordinator/internal/server/tcp"
	"goflix/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type WorkerStats struct {
	ID       string            `json:"id"`
	State    types.WorkerState `json:"state"`
	LastSeen time.Time         `json:"last_seen"`
	IP       string            `json:"ip"`
}

type SystemStats struct {
	// Process specific
	NumGoroutine int    `json:"num_goroutine"`
	Alloc        uint64 `json:"alloc_bytes"`
	Sys          uint64 `json:"sys_bytes"`
	NumGC        uint32 `json:"num_gc"`

	// System wide
	TotalRAM        uint64                 `json:"total_ram"`
	AvailableRAM    uint64                 `json:"available_ram"`
	UsedRAMPercent  float64                `json:"used_ram_percent"`
	TotalCPUCores   int                    `json:"total_cpu_cores"`
	CPUUsagePercent []float64              `json:"cpu_usage_percent"`
	CPUTemperatures []host.TemperatureStat `json:"cpu_temperatures"`
}

type MonitoringStatus struct {
	Timestamp time.Time     `json:"timestamp"`
	MongoDB   string        `json:"mongodb"`
	TCPServer string        `json:"tcp_server"`
	Workers   []WorkerStats `json:"workers"`
	System    SystemStats   `json:"system"`
}

type Service interface {
	GetStatus(ctx context.Context) MonitoringStatus
}

type monitoringService struct {
	mongoClient *plattform.MongoService
	tcpServer   *tcpserver.Server
}

func NewService(mongoClient *plattform.MongoService, tcpServer *tcpserver.Server) Service {
	return &monitoringService{
		mongoClient: mongoClient,
		tcpServer:   tcpServer,
	}
}

func (s *monitoringService) GetStatus(ctx context.Context) MonitoringStatus {
	// 1. MongoDB Status
	mongoStatus := "ok"
	if err := s.mongoClient.Ping(ctx); err != nil {
		mongoStatus = "down"
	}

	// 2. TCP Server & Workers
	s.tcpServer.Mu.RLock()
	workers := make([]WorkerStats, 0, len(s.tcpServer.Workers))
	for id, w := range s.tcpServer.Workers {
		ip := ""
		if w.Conn != nil {
			ip = w.Conn.RemoteAddr().String()
		}
		workers = append(workers, WorkerStats{
			ID:       id,
			State:    w.State,
			LastSeen: w.LastSeen,
			IP:       ip,
		})
	}
	s.tcpServer.Mu.RUnlock()

	// 3. System Stats (Process)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 4. System Stats (Host)
	vMem, _ := mem.VirtualMemory()
	cpuPercent, _ := cpu.Percent(0, true) // per cpu
	temps, _ := host.SensorsTemperatures()

	// Filter temps to try and find CPU ones if possible, or just return all
	// Usually keys are like "coretemp", "k10temp", etc.
	// We will return all found sensors for now.

	sysStats := SystemStats{
		NumGoroutine: runtime.NumGoroutine(),
		Alloc:        memStats.Alloc,
		Sys:          memStats.Sys,
		NumGC:        memStats.NumGC,

		TotalRAM:        0,
		AvailableRAM:    0,
		UsedRAMPercent:  0,
		TotalCPUCores:   runtime.NumCPU(),
		CPUUsagePercent: cpuPercent,
		CPUTemperatures: temps,
	}

	if vMem != nil {
		sysStats.TotalRAM = vMem.Total
		sysStats.AvailableRAM = vMem.Available
		sysStats.UsedRAMPercent = vMem.UsedPercent
	}

	return MonitoringStatus{
		Timestamp: time.Now(),
		MongoDB:   mongoStatus,
		TCPServer: "running",
		Workers:   workers,
		System:    sysStats,
	}
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/monitoring", h.GetMonitoringStatus)
}

func (h *Handler) GetMonitoringStatus(c *gin.Context) {
	status := h.svc.GetStatus(c.Request.Context())
	c.JSON(http.StatusOK, status)
}
