package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dataloader "goflix/api-coordinator/internal/data"
	"goflix/api-coordinator/internal/server/dispatcher"
	httpserver "goflix/api-coordinator/internal/server/http"
	tcpserver "goflix/api-coordinator/internal/server/tcp"
)

const defaultDispatchDelay = time.Minute / 2

type dispatchConfig struct {
	datasetPath  string
	targetUserID int
	delay        time.Duration
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := tcpserver.NewServer()
	go httpserver.NewRouter(ctx)

	results := make(chan dispatcher.Result, 32)
	resultTimeout := parseDurationEnv("DISPATCHER_RESULT_TIMEOUT", 90*time.Second)
	disp := dispatcher.New(server, results, resultTimeout)

	consumeResults(ctx, results)

	datasetPath := datasetPathFromEnv()
	targetUserID := parseEnvInt("TARGET_USER_ID", 0)
	delay := parseDurationEnv("AUTO_DISPATCH_DELAY", defaultDispatchDelay)
	if delay <= 0 {
		delay = defaultDispatchDelay
	}

	scheduleDatasetDispatch(ctx, disp, dispatchConfig{
		datasetPath:  datasetPath,
		targetUserID: targetUserID,
		delay:        delay,
	})

	log.Fatal(server.Start(os.Getenv("WORKER_TCP_ADDR")))
}

func consumeResults(ctx context.Context, results <-chan dispatcher.Result) {
	go func() {
		for {
			select {
			case res := <-results:
				log.Printf("[DISPATCHER] Resultado job=%s block=%d-%d vecinos=%d", res.JobID, res.BlockID.StartID, res.BlockID.EndID, len(res.Neighbors))
			case <-ctx.Done():
				return
			}
		}
	}()
}

func scheduleDatasetDispatch(ctx context.Context, disp *dispatcher.Dispatcher, cfg dispatchConfig) {
	go func() {
		select {
		case <-time.After(cfg.delay):
			log.Printf("[DISPATCHER] Leyendo dataset desde %s", cfg.datasetPath)
			userRatings, userIDs, err := dataloader.LoadUserRatings(cfg.datasetPath)
			if err != nil {
				log.Printf("[DISPATCHER] Error cargando dataset: %v", err)
				return
			}

			targetID := cfg.targetUserID
			if targetID == 0 {
				if len(userIDs) == 0 {
					log.Printf("[DISPATCHER] Dataset sin usuarios, no se despacha tarea")
					return
				}
				targetID = userIDs[0]
			}

			log.Printf("[DISPATCHER] Despachando tarea automática para userID=%d con %d usuarios", targetID, len(userRatings))
			disp.Run(ctx, targetID, userRatings)
		case <-ctx.Done():
			return
		}
	}()
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

func parseEnvInt(key string, fallback int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("[DISPATCHER] Valor inválido para %s: %s, usando %d", key, val, fallback)
		return fallback
	}
	return num
}
