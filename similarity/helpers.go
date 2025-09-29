// similarity/helpers.go
package similarity

import (
	"context"
	"math/rand"
	"time"
)

// BatchCollector genérico
type BatchCollector[T any] struct {
	In       <-chan T
	Out      chan<- []T
	Maxsize  int
	Duration time.Duration

	batch []T
	timer *time.Timer
}

func NewBatchCollector[T any](in <-chan T, out chan<- []T, maxSize int, duration time.Duration) *BatchCollector[T] {
	if maxSize <= 0 {
		panic("maxSize debe ser mayor que 0")
	}
	if duration <= 0 {
		panic("duration debe ser mayor que 0")
	}
	return &BatchCollector[T]{
		In:       in,
		Out:      out,
		Maxsize:  maxSize,
		Duration: duration,
		batch:    make([]T, 0, maxSize),
		timer:    time.NewTimer(duration),
	}
}

func (bc *BatchCollector[T]) flush() {
	if len(bc.batch) == 0 {
		return
	}
	b := make([]T, len(bc.batch))
	copy(b, bc.batch)
	bc.Out <- b
	bc.batch = bc.batch[:0]
}

func (bc *BatchCollector[T]) resetTimer() {
	if !bc.timer.Stop() {
		select {
		case <-bc.timer.C:
		default:
		}
	}
	bc.timer.Reset(bc.Duration)
}

func (bc *BatchCollector[T]) Run(ctx context.Context) {
	defer bc.timer.Stop()
	defer close(bc.Out)
	for {
		select {
		case <-ctx.Done():
			bc.flush()
			return
		case item, ok := <-bc.In:
			if !ok {
				bc.flush()
				return
			}
			bc.batch = append(bc.batch, item)
			if len(bc.batch) >= bc.Maxsize {
				bc.flush()
				bc.resetTimer()
			}
		case <-bc.timer.C:
			bc.flush()
			bc.resetTimer()
		}
	}
}

// Servicio Jaccard concurrente con batches
type Row struct {
	Set []uint32
}

type Matrix struct {
	Rows []Row
}

type Request struct {
	A     []uint32
	Reply chan []float64 // similitudes A vs todas las filas de M
	Ctx   context.Context
}

func CreateRandomFPMatrix(n, d int) [][]float64 {
	M := make([][]float64, n)
	for i := 0; i < n; i++ {
		M[i] = make([]float64, d)
		for j := 0; j < d; j++ {
			M[i][j] = rand.Float64()
		}
	}
	return M
}
