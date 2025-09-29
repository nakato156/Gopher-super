package similarity

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Query struct {
	A    []float64
	Resp chan []float64
}

type CosineServer struct {
	in   chan Query
	stop func()
}

func CosineWrapper(Mnorm [][]float64, batchSize int, timeout time.Duration) *CosineServer {
	in := make(chan Query, 10_000)
	out := make(chan []Query, 128)

	ctx, cancel := context.WithCancel(context.Background())
	collector := NewBatchCollector(in, out, batchSize, timeout)
	go collector.Run(ctx)

	// Un solo procesador de lotes que usa todos los cores adentro
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

// unroll de dot product por 8
func dotUnrolled8(a, b []float64) float64 {
	n := len(a)
	var s0 float64
	i := 0
	for ; i+7 < n; i += 8 {
		s0 += a[i+0]*b[i+0] +
			a[i+1]*b[i+1] +
			a[i+2]*b[i+2] +
			a[i+3]*b[i+3] +
			a[i+4]*b[i+4] +
			a[i+5]*b[i+5] +
			a[i+6]*b[i+6] +
			a[i+7]*b[i+7]
	}
	for ; i < n; i++ {
		s0 += a[i] * b[i]
	}
	return s0
}

// Computa el vector completo de cosenos
func runBatchCosines(batch []Query, Mnorm [][]float64) {
	if len(batch) == 0 {
		return
	}
	N := len(Mnorm)
	if N == 0 {
		for _, q := range batch {
			q.Resp <- nil
			close(q.Resp)
		}
		return
	}
	D := len(Mnorm[0])

	// Filtra queries inválidas sin romper B ni índices
	Q := make([][]float64, 0, len(batch))
	valid := batch[:0]
	for _, q := range batch {
		if len(q.A) != D {
			q.Resp <- nil
			close(q.Resp)
			continue
		}
		Q = append(Q, q.A)
		valid = append(valid, q)
	}
	batch = valid
	B := len(batch)
	if B == 0 {
		return
	}

	// Preasigna una matriz de resultados [B][N]
	scores := make([][]float64, B)
	for b := 0; b < B; b++ {
		scores[b] = make([]float64, N)
	}

	workers := runtime.NumCPU()

	// Tunea el tamaño de bloque
	// intenta dar ~3–4 bloques por worker
	perWorkerBlocks := 4
	block := (N + workers*perWorkerBlocks - 1) / (workers * perWorkerBlocks)
	if block < 512 {
		block = 512
	}
	if block > 4096 {
		block = 4096
	}

	type rng struct{ i0, i1 int }
	jobs := make(chan rng, workers*2)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Recorremos filas contiguas
				for i := job.i0; i < job.i1; i++ {
					row := Mnorm[i]
					// Para cada query del lote, hacemos el unroll dot
					for b := 0; b < B; b++ {
						s := dotUnrolled8(Q[b], row)
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

func TestCosineCon(M [][]float64, numUsers, vectorSize int) {
	rand.Seed(time.Now().UnixNano())

	batchSize := 256
	timeout := 5 * time.Millisecond
	server := CosineWrapper(M, batchSize, timeout)
	defer server.stop()

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

	fmt.Printf("Cosine Sim. NxN completada: %d usuarios, %d dimensiones\n", numUsers, vectorSize)
	fmt.Printf("Cosine Sim. Tiempo total concurrente: %v\n", elapsed)
}
