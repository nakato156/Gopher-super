package tcpserver

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"goflix/api-coordinator/internal/cache"
	"goflix/pkg/styles"
	"goflix/pkg/tcp"
	"goflix/pkg/types"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Worker struct {
	ID       string
	Conn     net.Conn
	State    types.WorkerState
	LastSeen time.Time
	SendCh   chan types.Message
}

// Server mantiene las conexiones activas y el canal central de entrada.
type Server struct {
	listener net.Listener
	Workers  map[string]*Worker
	Incoming chan types.Envelope
	mu       sync.RWMutex
}

const (
	redisWorkerIndexKey  = "workers:index"
	redisWorkerKeyPrefix = "worker:"
	redisWriteTimeout    = 2 * time.Second
	redisWorkerTTL       = 5 * time.Minute
)

var redisClient = cache.NewRedisClient()

// crea una instancia vacía del servidor TCP
func NewServer() *Server {
	return &Server{
		Workers:  make(map[string]*Worker),
		Incoming: make(chan types.Envelope, 100),
	}
}

func (s *Server) HandShake(conn net.Conn) (string, error) {
	// Implementar el handshake si es necesario

	worker_uuid := uuid.New()

	ackMsg := types.Message{
		Type: "ACK",
		Data: json.RawMessage(fmt.Sprintf(`{"worker_id":"%s"}`, worker_uuid.String())),
	}

	if err := tcp.WriteMessage(conn, ackMsg); err != nil {
		msg := fmt.Sprintf("[SERVER] Error enviando ACK:\n%v", err)
		styles.PrintFS("error", msg)
	}
	return worker_uuid.String(), nil
}

// Start abre el puerto TCP y empieza a aceptar conexiones entrantes de workers.
func (s *Server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("error al iniciar listener TCP: %w", err)
	}

	s.listener = ln
	msg := fmt.Sprintf("[SERVER] Escuchando en %s", addr)
	styles.PrintFS("default", msg)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			msg := fmt.Sprintf("[SERVER] Error al aceptar conexión:%v", err)
			styles.PrintFS("error", msg)
			continue
		}
		msg := fmt.Sprintf("[SERVER] Nueva conexión desde %s", conn.RemoteAddr())
		styles.PrintFS("info", msg)
		go s.handleConnection(conn)
	}
}

// handleConnection maneja una conexión de worker (lectura de mensajes)
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Leer longitud del mensaje (4 bytes big-endian)
		lengthBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lengthBuf)
		if err != nil {
			if err != io.EOF {
				msg := fmt.Sprintf("[SERVER] Error leyendo longitud:\n%v", err)
				styles.PrintFS("error", msg)
			}
			return
		}
		length := binary.BigEndian.Uint32(lengthBuf)

		// Leer cuerpo JSON
		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			msg := fmt.Sprintf("[SERVER] Error leyendo cuerpo:\n%v", err)
			styles.PrintFS("error", msg)
			return
		}

		var msg types.Message
		if err := json.Unmarshal(data, &msg); err != nil {

			msg := fmt.Sprintf("[SERVER] Error parseando JSON:%v", err)
			styles.PrintFS("error", msg)
			continue
		}

		// Extraer WorkerID si es un HELLO
		var workerID string
		if msg.Type == "HELLO" {
			var hello types.Hello
			if err := json.Unmarshal(msg.Data, &hello); err != nil {
				msg := fmt.Sprintf("[SERVER] Error parseando HELLO:%v", err)
				styles.PrintFS("error", msg)
				continue
			}

			workerID, err := s.HandShake(conn)
			if err != nil {
				fmsg := fmt.Sprintf("[SERVER] handshake error:\n%v", err)
				styles.PrintFS("error", fmsg)
				return
			}

			now := time.Now()
			// Registrar el worker
			s.mu.Lock()
			s.Workers[workerID] = &Worker{
				ID:       workerID,
				Conn:     conn,
				State:    types.WorkerIdle,
				LastSeen: now,
				SendCh:   make(chan types.Message, 10),
			}
			s.mu.Unlock()

			s.registerWorkerInRedis(workerID, hello.Concurrency, conn.RemoteAddr().String(), now)

			msg := fmt.Sprintf("[SERVER] Worker registrado: %s", workerID)
			styles.PrintFS("success", msg)
		}

		// Si el worker ya está en el mapa, usa su ID
		if workerID == "" {
			s.mu.RLock()
			for id, w := range s.Workers {
				if w.Conn == conn {
					workerID = id
					break
				}
			}
			s.mu.RUnlock()
		}

		if workerID == "" {
			styles.PrintFS("error", "[SERVER] Mensaje de worker desconocido")
			continue
		}

		// Enviar mensaje al canal global
		s.Incoming <- types.Envelope{
			WorkerID: workerID,
			Msg:      msg,
		}
	}
}

func (s *Server) registerWorkerInRedis(WorkerID string, Concurrency int, addr string, lastSeen time.Time) {
	if redisClient == nil || WorkerID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisWriteTimeout)
	defer cancel()

	fields := map[string]interface{}{
		"worker_id":   WorkerID,
		"concurrency": Concurrency,
		"state":       int(types.WorkerIdle),
		"last_seen":   lastSeen.UnixMilli(),
		"addr":        addr,
	}

	if err := redisClient.HSet(ctx, redisWorkerKeyPrefix+WorkerID, fields).Err(); err != nil {
		msg := fmt.Sprintf("[SERVER] Error registrando worker en Redis:\n%v", err)
		styles.PrintFS("error", msg)
		return
	}

	if err := redisClient.SAdd(ctx, redisWorkerIndexKey, WorkerID).Err(); err != nil {
		msg := fmt.Sprintf("[SERVER] Error indexando worker en Redis:\n%v", err)
		styles.PrintFS("error", msg)
		return
	}

	if err := redisClient.Expire(ctx, redisWorkerKeyPrefix+WorkerID, redisWorkerTTL).Err(); err != nil {
		msg := fmt.Sprintf("[SERVER] Error configurando TTL en Redis:\n%v", err)
		styles.PrintFS("error", msg)
	}
}
