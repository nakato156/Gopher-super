package dispatcher

import (
	"context"
	"goflix/api-coordinator/internal/tcpserver"
	"goflix/pkg/types"
	"time"
)

type (
	Task   = types.Task
	Result = types.Result
)

type Dispatcher struct {
	server   *tcpserver.Server
	blocks   <-chan Task   // producido por planner
	results  chan<- Result // para merger
	timeouts time.Duration
}

func (d *Dispatcher) Run(ctx context.Context) {
	// selecciona worker idle, envía TASK, marca Busy,
	// maneja reintentos si expira, procesa RESULT → results
}
