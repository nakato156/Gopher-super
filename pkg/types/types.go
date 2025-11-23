package types

import "encoding/json"

// WorkerState representa el estado actual de un worker.
type WorkerState int

const (
	WorkerIdle WorkerState = iota
	WorkerBusy
	WorkerDisconnected
)

// Message es el contenedor genérico que se envía por TCP.
// El campo Type indica el tipo de mensaje y Data contiene el payload serializado.
type Message struct {
	Type string          `json:"type"` // "HELLO","TASK","RESULT","HEARTBEAT","ERROR","ACK"
	Data json.RawMessage `json:"data"`
}

// ---- PAYLOADS ----

// Hello se envía cuando un worker se conecta al coordinador.
type Hello struct {
	WorkerID    string `json:"worker_id"`
	Concurrency int    `json:"concurrency"` // goroutines disponibles
}

// Block representa el rango (o partición) que debe procesar el worker.
type Block struct {
	StartID int `json:"start_id"`
	EndID   int `json:"end_id"`
}

// Task define una tarea enviada por el coordinador al worker.
type Task struct {
	JobID   string `json:"job_id"`
	BlockID Block  `json:"block_id"`
	// Algo             string                  `json:"algo"` // "user-based" | "item-based"
	// Sim              string                  `json:"sim"`  // "cosine" | "pearson" | "jaccard"
	K                int                     `json:"k"`
	TargetRatings    map[int]float64         `json:"target_ratings"`
	CandidateRatings map[int]map[int]float64 `json:"candidate_ratings"`
}

// Neighbor representa una relación de similitud parcial (resultado intermedio).
type Neighbor struct {
	ID         string  `json:"id"`
	Similarity float64 `json:"similarity"`
}

// Result es la respuesta que el worker envía al coordinador tras procesar un bloque.
type Result struct {
	JobID     string     `json:"job_id"`
	BlockID   Block      `json:"block_id"`
	Neighbors []Neighbor `json:"neighbors"`
}

// Heartbeat mantiene viva la conexión y reporta estado del worker.
type Heartbeat struct {
	WorkerID string  `json:"worker_id"`
	Busy     bool    `json:"busy"`
	CPU      float64 `json:"cpu"` // opcional, uso de CPU o carga
}

// Envelope asocia un mensaje recibido con el ID del worker que lo envió.
// Es útil en el coordinador para enrutar y manejar mensajes sin perder contexto.
type Envelope struct {
	WorkerID string
	Msg      Message
}

// HTTP
type UserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}
