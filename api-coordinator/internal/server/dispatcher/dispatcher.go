package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	tcpserver "goflix/api-coordinator/internal/server/tcp"
	"goflix/pkg/types"
)

type (
	Task   = types.Task
	Result = types.Result
)

type Dispatcher struct {
	server       *tcpserver.Server
	timeouts     time.Duration
	resultsChans map[string]chan types.Result
	mu           sync.Mutex
}

func New(server *tcpserver.Server, timeout time.Duration) *Dispatcher {
	d := &Dispatcher{
		server:       server,
		timeouts:     timeout,
		resultsChans: make(map[string]chan types.Result),
	}
	go d.processIncoming()
	return d
}

func (d *Dispatcher) Run(ctx context.Context, userID int, userRatings map[int]map[int]float64, resultsCh chan<- Result) (int, error) {
	fmt.Println("Dispatcher Run started for userID:", userID)
	if d.resultsChans == nil {
		d.resultsChans = make(map[string]chan types.Result)
	}

	d.server.Mu.RLock()
	idleWorkers := make([]string, 0)
	for id, w := range d.server.Workers {
		if w.State == types.WorkerIdle {
			idleWorkers = append(idleWorkers, id)
		}
	}
	d.server.Mu.RUnlock()
	fmt.Println("Idle workers found:", len(idleWorkers))

	if len(idleWorkers) == 0 {
		fmt.Println("No idle workers, returning")
		return 0, nil
	}

	userIDs := make([]int, 0, len(userRatings)-1)
	for id := range userRatings {
		if id != userID {
			userIDs = append(userIDs, id)
		}
	}
	sort.Ints(userIDs)

	numBlocks := len(idleWorkers)
	if numBlocks > len(userIDs) {
		numBlocks = len(userIDs)
		idleWorkers = idleWorkers[:numBlocks]
	}
	blockSize := len(userIDs) / numBlocks
	remainder := len(userIDs) % numBlocks
	fmt.Println("Particionando", len(userIDs), "usuarios en", numBlocks, "bloques")

	// jobID := uuid.New().String()
	startIdx := 0
	dispatchedCount := 0
	for i := 0; i < numBlocks; i++ {
		endIdx := startIdx + blockSize
		if i < remainder {
			endIdx++
		}
		workerID := idleWorkers[i]
		task := types.Task{
			JobID:   workerID,
			BlockID: types.Block{StartID: startIdx, EndID: endIdx - 1},
			K:       30,
		}
		fmt.Println("Creando tarea para worker", workerID, "con block StartID:", startIdx, "EndID:", endIdx-1)
		ch := make(chan types.Result, 1)
		d.mu.Lock()
		d.resultsChans[workerID] = ch
		d.mu.Unlock()

		// mandar tarea al worker
		err := d.DispatchTask(workerID, task)
		if err != nil {
			fmt.Println("Error dispatching task to worker", workerID, ":", err)
			d.mu.Lock()
			delete(d.resultsChans, workerID)
			d.mu.Unlock()
			continue
		}

		// establaecer como ocupados
		d.server.Mu.Lock()
		d.server.Workers[workerID].State = types.WorkerBusy
		d.server.Mu.Unlock()
		fmt.Println("Worker", workerID, "marcado como busy")

		go func(wID string, c chan types.Result, jID string) {
			fmt.Println("Esperando resultado para job", jID, "en worker", wID)
			select {
			case res := <-c:
				fmt.Println("Recibido resultado para job", jID, "en worker", wID)
				resultsCh <- res
				d.server.Mu.Lock()
				d.server.Workers[wID].State = types.WorkerIdle
				d.server.Mu.Unlock()
				fmt.Println("Worker", wID, "marcado como idle")
			case <-time.After(d.timeouts):
				fmt.Println("Timeout esperando resultado para job", jID, "en worker", wID)
				d.server.Mu.Lock()
				d.server.Workers[wID].State = types.WorkerIdle
				d.server.Mu.Unlock()
			}
			d.mu.Lock()
			delete(d.resultsChans, jID)
			d.mu.Unlock()
		}(workerID, ch, workerID)

		startIdx = endIdx
		dispatchedCount++
	}
	return dispatchedCount, nil
}

func (d *Dispatcher) processIncoming() {
	// procesa los mensajes de los workers, para cuando ya retornan los resultados
	// y también gestiona los heartbeats para monitorear el estado
	for env := range d.server.Incoming {
		switch env.Msg.Type {
		case "RESULT":
			var result types.Result
			if err := json.Unmarshal(env.Msg.Data, &result); err != nil {
				continue
			}
			fmt.Println("Procesando RESULT para job", result.JobID, "de worker", env.WorkerID)
			d.mu.Lock()
			ch, ok := d.resultsChans[env.WorkerID]
			d.mu.Unlock()
			if ok {
				select {
				case ch <- result:
				default:
				}
			}
		}
	}
}

func (d *Dispatcher) DispatchTask(WorkerID string, task types.Task) (err error) {
	// manda la tarea a un worker especifico
	fmt.Println("Dispatching task to worker", WorkerID)
	d.server.Mu.RLock()
	worker, ok := d.server.Workers[WorkerID]
	d.server.Mu.RUnlock()
	if !ok {
		return fmt.Errorf("worker not found")
	}
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	msg := types.Message{
		Type: "TASK",
		Data: data,
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker %s channel closed", WorkerID)
		}
	}()
	select {
	case worker.SendCh <- msg:
		return nil
	default:
		return fmt.Errorf("send channel full")
	}
}
