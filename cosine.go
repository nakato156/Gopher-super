package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func normalizeRow(v []float64) {
	var s float64
	for _, x := range v {
		s += x * x
	}
	if s == 0 {
		return
	}
	inv := 1.0 / math.Sqrt(s)
	for i := range v {
		v[i] *= inv
	}
}

func normalizeRows(M [][]float64) {
	for i := range M {
		normalizeRow(M[i])
	}
}

func dot(a, b []float64) float64 {
	var s float64
	for i := 0; i < len(a); i++ {
		s += a[i] * b[i]
	}
	return s
}

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

type Query struct {
	A    []float64
	Resp chan []float64
}

type CosineServer struct {
	in   chan Query
	stop func()
}

func StartCosineServer(Mnorm [][]float64, batchSize int, timeout time.Duration) *CosineServer {
	in := make(chan Query, 10_000)
	out := make(chan []Query, 128)

	ctx, cancel := context.WithCancel(context.Background())
	collector := NewBatchCollector(in, out, batchSize, timeout)
	go collector.Run(ctx)

	// n solo nivel de paralelismo dentro
	go func() {
		for batch := range out {
			runBatchCosines(batch, Mnorm)
		}
	}()

	return &CosineServer{
		in:   in,
		stop: func() { cancel(); close(in) },
	}
}

// Computa el vector completo de cosenos
func runBatchCosines(batch []Query, Mnorm [][]float64) {
	B := len(batch)
	if B == 0 {
		return
	}
	N := len(Mnorm)
	D := len(Mnorm[0])

	// Copias de A
	Q := make([][]float64, B)
	for i, q := range batch {
		if len(q.A) != D {
			// responde vacío si hay mismatch
			q.Resp <- nil
			close(q.Resp)
			batch[i] = batch[len(batch)-1]
			batch = batch[:len(batch)-1]
			i--
			continue
		}
		Q[i] = q.A
	}

	// scores[b][i] = dot(Q[b], Mnorm[i])
	scores := make([][]float64, B)
	for b := 0; b < B; b++ {
		scores[b] = make([]float64, N)
	}

	// Pool tamaño NumCPU; paralelismo por bloques de filas de M
	workers := runtime.NumCPU()
	const block = 256

	type rng struct{ i0, i1 int }
	jobs := make(chan rng, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				for i := job.i0; i < job.i1; i++ {
					row := Mnorm[i]
					// calcular para todas las queries del lote
					for b := 0; b < B; b++ {
						s := 0.0
						qb := Q[b]
						// dot sin sqrt: M y A ya normalizados
						for d := 0; d < D; d++ {
							s += qb[d] * row[d]
						}
						scores[b][i] = s
					}
				}
			}
		}()
	}
	for i0 := 0; i0 < N; i0 += block {
		i1 := i0 + block
		if i1 > N {
			i1 = N
		}
		jobs <- rng{i0, i1}
	}
	close(jobs)
	wg.Wait()

	for b, q := range batch {
		q.Resp <- scores[b]
		close(q.Resp)
	}
}

func createRandomMatrix(n, d int) [][]float64 {
	M := make([][]float64, n)
	for i := 0; i < n; i++ {
		M[i] = make([]float64, d)
		for j := 0; j < d; j++ {
			M[i][j] = rand.Float64() // 0.124721  0.124722
		}
	}
	return M
}

func main() {
	rand.Seed(time.Now().UnixNano())

	numUsers := 8 * 4
	vectorSize := 1024 * 2

	M := createRandomMatrix(numUsers, vectorSize)
	// normalizeRows(M) // cos = dot(A, M[i])

	batchSize := 64 * 4
	timeout := 5 * time.Millisecond
	server := StartCosineServer(M, batchSize, timeout)
	defer server.stop()

	// Matriz de resultados NxN
	results := make([][]float64, numUsers)
	for i := 0; i < numUsers; i++ {
		results[i] = make([]float64, numUsers)
	}

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(numUsers)

	// Emular 10k usuarios concurrentes
	for u := 0; u < numUsers; u++ {
		go func(u int) {
			defer wg.Done()
			resp := make(chan []float64, 1)
			server.in <- Query{A: M[u], Resp: resp}
			sims := <-resp
			results[u] = sims
		}(u)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Similitud NxN completada: %d usuarios, %d dimensiones\n", numUsers, vectorSize)
	fmt.Printf("batchSize=%d timeout=%v workers=%d\n", batchSize, timeout, runtime.NumCPU())
	fmt.Printf("Tiempo total concurrente (batching global): %v\n", elapsed)
	fmt.Printf("Ejemplo: sim[0][1]=%.6f  sim[1][0]=%.6f\n", results[0][1], results[1][0])
}
