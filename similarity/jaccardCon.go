package similarity

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func quickSortUint32(a []uint32) {
	n := len(a)
	if n < 2 {
		return
	}
	type seg struct{ l, r int }
	stack := make([]seg, 0, 64)
	stack = append(stack, seg{0, n - 1})

	for len(stack) > 0 {
		// pop
		l := stack[len(stack)-1].l
		r := stack[len(stack)-1].r
		stack = stack[:len(stack)-1]

		for l < r {
			i, j := l, r
			mid := l + (r-l)/2
			if a[mid] < a[l] {
				a[l], a[mid] = a[mid], a[l]
			}
			if a[r] < a[l] {
				a[l], a[r] = a[r], a[l]
			}
			if a[r] < a[mid] {
				a[mid], a[r] = a[r], a[mid]
			}
			pivot := a[mid]

			for i <= j {
				for a[i] < pivot {
					i++
				}
				for a[j] > pivot {
					j--
				}
				if i <= j {
					a[i], a[j] = a[j], a[i]
					i++
					j--
				}
			}

			// Procesa primero el subarreglo más pequeño (el otro va a la pila)
			if (j - l) < (r - i) {
				if i < r {
					stack = append(stack, seg{i, r})
				}
				r = j
			} else {
				if l < j {
					stack = append(stack, seg{l, j})
				}
				l = i
			}
		}
	}
}

// compacta valores únicos in-place y devuelve el slice reducido.
func dedupSorted(a []uint32) []uint32 {
	if len(a) == 0 {
		return a
	}
	w := 1
	for i := 1; i < len(a); i++ {
		if a[i] != a[w-1] {
			a[w] = a[i]
			w++
		}
	}
	return a[:w]
}

// ordena con quicksort y deduplica.
func uniqueSorted(a []uint32) []uint32 {
	if len(a) == 0 {
		return a
	}
	quickSortUint32(a)
	return dedupSorted(a)
}

func jaccardSorted(Auniq, Buniq []uint32) float64 {
	la := len(Auniq)
	lb := len(Buniq)

	// Casos triviales
	if la == 0 && lb == 0 {
		return 1.0
	}
	if la == 0 || lb == 0 {
		return 0.0
	}

	// Heurística para elegir galloping cuando hay gran desbalance
	const ratio = 8

	// Helpers locales (sin agregar funciones globales)
	lowerBound := func(arr []uint32, lo, hi int, x uint32) int {
		for lo < hi {
			m := lo + (hi-lo)/2
			if arr[m] < x {
				lo = m + 1
			} else {
				hi = m
			}
		}
		return lo
	}
	gallopFind := func(arr []uint32, start int, x uint32) (pos int, found bool) {
		n := len(arr)
		if start >= n {
			return n, false
		}
		lo := start
		hi := lo + 1
		for hi < n && arr[hi] < x {
			lo = hi
			step := hi - start + 1
			hi += step
		}
		if hi > n {
			hi = n
		}
		i := lowerBound(arr, lo, hi, x)
		if i < n && arr[i] == x {
			return i + 1, true
		}
		return i, false
	}

	// Caso optimizado: tamaños iguales => recorrido casi branchless + unroll 2×
	if la == lb {
		i, j, inter := 0, 0, 0
		n := la

		// Unroll 2×
		for i+1 < n && j+1 < n {
			ai := Auniq[i]
			bj := Buniq[j]
			if ai <= bj {
				i++
			}
			if bj <= ai {
				j++
			}
			if ai == bj {
				inter++
			}

			ai = Auniq[i]
			bj = Buniq[j]
			if ai <= bj {
				i++
			}
			if bj <= ai {
				j++
			}
			if ai == bj {
				inter++
			}
		}
		// Resto
		for i < n && j < n {
			ai := Auniq[i]
			bj := Buniq[j]
			if ai <= bj {
				i++
			}
			if bj <= ai {
				j++
			}
			if ai == bj {
				inter++
			}
		}

		union := 2*n - inter
		if union == 0 {
			return 1.0
		}
		return float64(inter) / float64(union)
	}

	// Caminos híbridos (galloping) cuando hay fuerte desbalance
	inter := 0
	if la*ratio < lb {
		j := 0
		for i := 0; i < la; i++ {
			pos, ok := gallopFind(Buniq, j, Auniq[i])
			if ok {
				inter++
			}
			j = pos
			if j >= lb {
				break
			}
		}
	} else if lb*ratio < la {
		i := 0
		for j := 0; j < lb; j++ {
			pos, ok := gallopFind(Auniq, i, Buniq[j])
			if ok {
				inter++
			}
			i = pos
			if i >= la {
				break
			}
		}
	} else {
		// Dos punteros clásico
		i, j := 0, 0
		for i < la && j < lb {
			ai, bj := Auniq[i], Buniq[j]
			if ai == bj {
				inter++
				i++
				j++
			} else if ai < bj {
				i++
			} else {
				j++
			}
		}
	}

	union := la + lb - inter
	if union == 0 {
		return 1.0
	}
	return float64(inter) / float64(union)
}

func NewService(M *Matrix, workers, block, maxBatch int, dur time.Duration) *Service {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if workers <= 0 {
		workers = 3 * runtime.NumCPU()
	}
	if block <= 0 {
		block = 32
	}
	in := make(chan Request, 8192)
	out := make(chan []Request, 256)
	bc := NewBatchCollector[Request](in, out, maxBatch, dur)
	s := &Service{M: M, In: in, OutBatch: out, Workers: workers, Block: block}
	go bc.Run(context.Background())
	go s.loopBatches()
	return s
}

// el caller manda A crudo
// se copia, ordena y deduplica
func (s *Service) Submit(ctx context.Context, A []uint32) ([]float64, bool) {
	cpy := make([]uint32, len(A))
	copy(cpy, A)
	Auniq := uniqueSorted(cpy)
	rep := make(chan []float64, 1)
	req := Request{A: Auniq, Reply: rep, Ctx: ctx}
	select {
	case s.In <- req:
	case <-ctx.Done():
		return nil, false
	}
	select {
	case res := <-rep:
		return res, true
	case <-ctx.Done():
		return nil, false
	}
}

func (s *Service) loopBatches() {
	for batch := range s.OutBatch {
		s.processBatch(batch)
	}
}

func (s *Service) processBatch(batch []Request) {
	if len(batch) == 0 {
		return
	}
	rows := s.M.Rows
	outs := make([][]float64, len(batch))
	for i := range batch {
		outs[i] = make([]float64, len(rows))
	}

	// tareas por bloque de filas (mejor localidad)
	// Cada worker toma un bloque [lo,hi) de filas y, para cada request del batch,
	// calcula outs[i][r]
	// Así, la fila r se lee una sola vez para todos los A del batch
	type task struct {
		lo int
		hi int
	}
	jobs := make(chan task, s.Workers*32)

	var wg sync.WaitGroup
	wg.Add(s.Workers)
	for w := 0; w < s.Workers; w++ {
		go func() {
			defer wg.Done()
			for t := range jobs {
				for r := t.lo; r < t.hi; r++ {
					rowSet := rows[r].Set
					for i := range batch {
						select {
						case <-batch[i].Ctx.Done():
							continue
						default:
						}
						outs[i][r] = jaccardSorted(batch[i].A, rowSet)
					}
				}
			}
		}()
	}

	blk := s.Block
	if blk < 16 {
		blk = 16
	}
	if blk > len(rows) {
		blk = len(rows)
	}
	for start := 0; start < len(rows); start += blk {
		end := start + blk
		if end > len(rows) {
			end = len(rows)
		}
		jobs <- task{lo: start, hi: end}
	}
	close(jobs)
	wg.Wait()

	// Responder a cada request
	for i, req := range batch {
		select {
		case <-req.Ctx.Done():
		default:
			req.Reply <- outs[i]
		}
		close(req.Reply)
	}
}

func BuildMatrix(nRows, rowSize, universe int) *Matrix {
	rows := make([]Row, nRows)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < nRows; i++ {
		tmp := make([]uint32, rowSize)
		for j := 0; j < rowSize; j++ {
			tmp[j] = uint32(rng.Intn(universe))
		}
		rows[i] = Row{Set: uniqueSorted(tmp)}
	}
	return &Matrix{Rows: rows}
}

func BuildAVector(rowSize int) *[]uint32 {
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + 1337))
	A := make([]uint32, rowSize)
	for j := 0; j < rowSize; j++ {
		A[j] = uint32(rng.Intn(2000))
	}
	return &A
}

func TestJaccardCon(M *Matrix, A []uint32, batch int, workers int, nRequests int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	svc := NewService(M, workers, 32, batch, 3*time.Millisecond)

	startTime := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < nRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			_, ok := svc.Submit(ctx, A)
			if !ok {
				fmt.Printf("Request %d cancelado\n", id)
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("Jaccard Concurrente Tiempo total: %s\n", elapsed)
}
