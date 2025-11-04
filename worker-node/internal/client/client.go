package client

import (
	"encoding/json"
	"errors"
	"goflix/pkg/tcp"
	"goflix/pkg/types"
	"net"
	"runtime"
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
