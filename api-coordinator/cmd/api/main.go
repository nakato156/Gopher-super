package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	dataloader "goflix/api-coordinator/internal/data"
	"goflix/api-coordinator/internal/server/dispatcher"
	httpserver "goflix/api-coordinator/internal/server/http"
	tcpserver "goflix/api-coordinator/internal/server/tcp"
)

type dispatchData struct {
	userID      int
	userRatings map[int]map[int]float64
	userIDs     []int
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := tcpserver.NewServer()

	resultTimeout := parseDurationEnv("DISPATCHER_RESULT_TIMEOUT", 90*time.Second)
	disp := dispatcher.New(server, resultTimeout)

	datasetPath := datasetPathFromEnv()

	log.Printf("[SERVER] Leyendo dataset desde %s", datasetPath)
	userRatings, userIDs, err := dataloader.LoadUserRatings(datasetPath)
	var mu sync.RWMutex
	if err != nil {
		log.Printf("[SERVER] Error cargando dataset: %v", err)
		return
	}

	triggerDispatch := func(userID int) ([]dispatcher.Result, error) {
		payload := dispatchData{
			userID:      userID,
			userRatings: userRatings,
			userIDs:     userIDs,
		}
		return scheduleDatasetDispatch(ctx, disp, payload, &mu)
	}

	go httpserver.NewRouter(ctx, triggerDispatch)

	log.Fatal(server.Start(os.Getenv("WORKER_TCP_ADDR")))
}

func scheduleDatasetDispatch(ctx context.Context, disp *dispatcher.Dispatcher, data dispatchData, mu *sync.RWMutex) ([]dispatcher.Result, error) {
	mu.RLock()
	defer mu.RUnlock()
	targetID := data.userID

	if targetID == 0 {
		if len(data.userIDs) == 0 {
			log.Printf("[DISPATCHER] Dataset sin usuarios, no se despacha tarea")
			return nil, nil
		}
		targetID = data.userIDs[0]
	}

	log.Printf("[DISPATCHER] Despachando tarea automática para userID=%d con %d usuarios", targetID, len(data.userRatings))

	resultsCh := make(chan dispatcher.Result, 100) // Buffer sufficiente para evitar bloqueo
	count, err := disp.Run(ctx, targetID, data.userRatings, resultsCh)
	if err != nil {
		return nil, err
	}

	var results []dispatcher.Result
	for i := 0; i < count; i++ {
		select {
		case res := <-resultsCh:
			results = append(results, res)
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}
	return results, nil
}

func datasetPathFromEnv() string {
	if val := strings.TrimSpace(os.Getenv("RATINGS_DATA_PATH")); val != "" {
		return val
	}
	return filepath.Join("dataset", "ml-latest-small", "ratings.csv")
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(val); err == nil {
		return time.Duration(secs) * time.Second
	}
	log.Printf("[DISPATCHER] Valor inválido para %s: %s, usando %v", key, val, fallback)
	return fallback
}
