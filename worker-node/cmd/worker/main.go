package main

import (
	"context"
	"errors"
	"fmt"
	"goflix/pkg/styles"
	"goflix/worker-node/internal/client"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const heartbeatInterval = 5 * time.Second

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	worker := client.NewClient()
	worker.State = client.StConnecting

	coordinatorAddr := os.Getenv("COORDINATOR_ADDR")
	if coordinatorAddr == "" {
		styles.PrintFS("error", "[WORKER] Variable COORDINATOR_ADDR no definida")
		return
	}

	conn, err := net.Dial("tcp", coordinatorAddr)
	if err != nil {
		worker.State = client.StDisconnected
		styles.PrintFS("error", fmt.Sprintf("[WORKER] Error de conexión: %v", err))
		return
	}

	styles.PrintFS("success", "[WORKER] Conexión TCP establecida con el coordinador")

	defer conn.Close()

	worker.Conn = conn
	worker.State = client.StHandshaking

	workerID, err := worker.HandShake(conn)
	if err != nil {
		styles.PrintFS("error", fmt.Sprintf("[WORKER] Error en el handshake: %v", err))
		return
	}

	styles.PrintFS("success", fmt.Sprintf("[WORKER] Handshake completado. Worker ID asignado: %s", workerID))

	heartbeatDone := make(chan error, 1)
	taskDone := make(chan error, 1)

	go func() {
		heartbeatDone <- worker.StartHeartbeat(ctx, heartbeatInterval)
	}()

	go func() {
		taskDone <- worker.Process(ctx)
	}()

	select {
	case err := <-heartbeatDone:
		if err != nil && !errors.Is(err, context.Canceled) {
			styles.PrintFS("error", fmt.Sprintf("[WORKER] Heartbeat detenido: %v", err))
		}
	case <-ctx.Done():
		cancel()
		if err := <-heartbeatDone; err != nil && !errors.Is(err, context.Canceled) {
			styles.PrintFS("error", fmt.Sprintf("[WORKER] Heartbeat detenido: %v", err))
		}
		styles.PrintFS("info", "[WORKER] Señal recibida, cerrando worker")
	}
}
