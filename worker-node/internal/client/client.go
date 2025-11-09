package client

import (
	"context"
	"encoding/json"
	"errors"
	"goflix/pkg/tcp"
	"goflix/pkg/types"
	"net"
	"runtime"
	"sync"
	"time"
)

type ClientState int

const (
	StDisconnected ClientState = iota
	StConnecting
	StHandshaking
	StReady
	StWorking
	StShuttingDown
)

type WorkerClient struct {
	ID          string
	Conn        net.Conn
	State       ClientState // estado del ciclo de vida del cliente
	Busy        bool        // ocupación local (equivalente a “idle/busy”)
	CurrentTask *types.Task // nil si no hay trabajo
	LastSeen    time.Time   // para métricas/timeouts
	connMu      sync.Mutex
}

var payload struct {
	WorkerID string `json:"worker_id"`
}

func NewClient() *WorkerClient {
	return &WorkerClient{
		ID:          "",
		Conn:        nil,
		State:       StDisconnected,
		Busy:        false,
		CurrentTask: nil,
		LastSeen:    time.Now(),
	}
}

func (wc *WorkerClient) HandShake(conn net.Conn) (string, error) {
	wc.State = StHandshaking

	// Enviar HELLO (ID vacío, el server lo asigna)
	hello := types.Hello{
		WorkerID:    "", // el server lo da
		Concurrency: runtime.NumCPU(),
	}
	data, _ := json.Marshal(hello)
	msg := types.Message{Type: "HELLO", Data: data}

	if err := tcp.WriteMessage(conn, msg); err != nil {
		wc.State = StDisconnected
		return "", err
	}

	// Esperar ACK con un timeout razonable
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetReadDeadline(time.Time{}) // limpia el deadline

	ack, err := tcp.ReadMessage(conn)
	if err != nil {
		wc.State = StDisconnected
		return "", err
	}
	if ack.Type != "ACK" {
		wc.State = StDisconnected
		return "", errors.New("handshake: esperaba ACK del servidor")
	}

	// Parsear el worker_id devuelto
	if err := json.Unmarshal(ack.Data, &payload); err != nil {
		wc.State = StDisconnected
		return "", err
	}
	if payload.WorkerID == "" {
		wc.State = StDisconnected
		return "", errors.New("handshake: ACK sin worker_id")
	}

	// Actualizar estado del cliente
	wc.ID = payload.WorkerID
	wc.Conn = conn
	wc.State = StReady
	wc.LastSeen = time.Now()
	return wc.ID, nil
}

func (wc *WorkerClient) sendMessage(msg types.Message) error {
	wc.connMu.Lock()
	defer wc.connMu.Unlock()

	if wc.Conn == nil {
		return errors.New("worker client: conexión no inicializada")
	}

	return tcp.WriteMessage(wc.Conn, msg)
}

func (wc *WorkerClient) StartHeartbeat(ctx context.Context, interval time.Duration) error {
	if wc.ID == "" {
		return errors.New("heartbeat: worker sin ID asignado")
	}

	if wc.Conn == nil {
		return errors.New("heartbeat: conexión no disponible")
	}

	if interval <= 0 {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			hb := types.Heartbeat{
				WorkerID: wc.ID,
				Busy:     wc.Busy,
				CPU:      0,
			}

			data, err := json.Marshal(hb)
			if err != nil {
				return err
			}

			msg := types.Message{Type: "HEARTBEAT", Data: data}

			if err := wc.sendMessage(msg); err != nil {
				wc.State = StDisconnected
				return err
			}

			wc.LastSeen = time.Now()
		}
	}
}

func (wc *WorkerClient) Process(ctx context.Context) error {
	if wc.ID == "" {
		return errors.New("Worker sin ID asignado")
	}

	if wc.Conn == nil {
		return errors.New("Worker sin conexion")
	}
	// calculo de cosine similarity
	return nil
}
