package main

import (
	"fmt"
	"goflix/pkg/styles"
	"goflix/worker-node/internal/client"
	"net"
	"os"
)

func main() {
	worker := client.NewClient()
	worker.State = client.StConnecting

	conn, err := net.Dial("tcp", os.Getenv("COORDINATOR_ADDR"))

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
}
