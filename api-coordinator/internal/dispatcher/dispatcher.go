package dispatcher

import (
	"context"
	"time"

	"api-coordinator/internal/server"
	"api-coordinator/pkg/types"
)

type (
	Task   = types.Task
	Result = types.Result
)

type Dispatcher struct {
	server   *server.Server
	blocks   <-chan Task   // producido por planner
	results  chan<- Result // para merger
	timeouts time.Duration
}

func (d *Dispatcher) Run(ctx context.Context) {
	// selecciona worker idle, envía TASK, marca Busy,
	// maneja reintentos si expira, procesa RESULT → results
}
