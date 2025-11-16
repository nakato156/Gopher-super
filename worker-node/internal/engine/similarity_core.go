package engine

import (
	"container/heap"
	"goflix/worker-node/internal/types"
	"math"
	"runtime"
	"sort"
	"sync"
)

func makeKey(i, j int) uint64 {
	if i > j {
		i, j = j, i
	}
	return (uint64(uint32(i)) << 32) | uint64(uint32(j))
}
func splitKey(k uint64) (int, int) {
	i := int(int32(k >> 32))
	j := int(int32(k & 0xffffffff))
	return i, j
}

type partialAcc struct {
	dot  map[uint64]float64 // producto punto
	co   map[uint64]int     // co-ratings
	norm map[int]float64    // ||i||^2
}

func newPartialAcc() *partialAcc {
	return &partialAcc{
		dot:  make(map[uint64]float64, 1<<12),
		co:   make(map[uint64]int, 1<<12),
		norm: make(map[int]float64, 1<<12),
	}
}

func (p *partialAcc) addDot(i, j int, v float64) {
	k := makeKey(i, j)
	p.dot[k] += v
	p.co[k]++
}

func (p *partialAcc) addNorm(i int, v float64) {
	p.norm[i] += v
}

func mergeAcc(dst, src *partialAcc) {
	for k, v := range src.dot {
		dst.dot[k] += v
	}
	for k, c := range src.co {
		dst.co[k] += c
	}
	for i, v := range src.norm {
		dst.norm[i] += v
	}
}

// Procesa un bloque de usuarios y acumula en un partialAcc dado. Reutiliza buf para evitar allocs por usuario.
func processUserBlock(acc *partialAcc, users []int, userRatings map[int]map[int]float64, buf []int) []int {
	for _, u := range users {
		ru := userRatings[u]
		if len(ru) == 0 {
			continue
		}

		// normas por ítem (sum r^2)
		for i, r := range ru {
			acc.addNorm(i, r*r)
		}

		// items del usuario en buf[0:n]
		n := 0
		for i := range ru {
			if n < len(buf) {
				buf[n] = i
			} else {
				buf = append(buf, i)
			}
			n++
		}

		// pares de ítems co-calificados por este usuario
		for a := 0; a < n; a++ {
			i := buf[a]
			ri := ru[i]
			for b := a + 1; b < n; b++ {
				j := buf[b]
				rj := ru[j]
				acc.addDot(i, j, ri*rj)
			}
		}
	}
	return buf
}

func chunkInts(all []int, size int) [][]int {
	var out [][]int
	for i := 0; i < len(all); i += size {
		j := i + size
		if j > len(all) {
			j = len(all)
		}
		out = append(out, all[i:j])
	}
	return out
}

type SimRow = types.SimRow
type SimMatrix = types.SimMatrix
type NeighborList = types.NeighborList

func BuildSimilaritiesConcurrent(userRatings map[int]map[int]float64, workerCount int) (types.SimMatrix, map[int]float64) {
	if workerCount < 1 {
		workerCount = 1
	}

	users := make([]int, 0, len(userRatings))
	for u := range userRatings {
		users = append(users, u)
	}
	sort.Ints(users)

	cpus := runtime.NumCPU()
	targetBlocks := cpus * 8
	if targetBlocks < workerCount {
		targetBlocks = workerCount
	}
	if targetBlocks < 16 {
		targetBlocks = 16
	}
	batch := (len(users) + targetBlocks - 1) / targetBlocks
	if batch < 10 {
		batch = 10
	}
	blocks := chunkInts(users, batch)

	// Lanza workers con su partialAcc local
	workCh := make(chan []int)
	partCh := make(chan *partialAcc, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for w := 0; w < workerCount; w++ {
		go func() {
			defer wg.Done()
			local := newPartialAcc()
			buf := make([]int, 0, 64)
			for blk := range workCh {
				buf = processUserBlock(local, blk, userRatings, buf)
			}
			partCh <- local
		}()
	}

	go func() {
		for _, blk := range blocks {
			workCh <- blk
		}
		close(workCh)
		wg.Wait()
		close(partCh)
	}()

	// Merge de todos los parciales en uno global
	global := newPartialAcc()
	for p := range partCh {
		mergeAcc(global, p)
	}

	// Precomputar normas (sqrt)
	norms := make(map[int]float64, len(global.norm))
	for i, sum := range global.norm {
		if sum > 0 {
			norms[i] = math.Sqrt(sum)
		}
	}

	// Construir SimMatrix secuencialmente
	sim := make(SimMatrix, 1<<12)

	for k, dot := range global.dot {
		co := global.co[k]
		if co < types.MinCoRatings {
			continue
		}
		i, j := splitKey(k)
		ni := norms[i]
		nj := norms[j]
		if ni == 0 || nj == 0 {
			continue
		}
		val := dot / (ni * nj)
		if val == 0 {
			continue
		}

		rowI := sim[i]
		if rowI == nil {
			rowI = make(SimRow)
			sim[i] = rowI
		}
		rowI[j] = val

		rowJ := sim[j]
		if rowJ == nil {
			rowJ = make(SimRow)
			sim[j] = rowJ
		}
		rowJ[i] = val
	}

	return sim, norms
}

// top-K vecinos por ítem usando heap mínimo
func TopKNeighbors(sim types.SimMatrix, k int) types.NeighborList {
	out := make(types.NeighborList, len(sim))
	for i, row := range sim {
		h := &types.NeighborHeap{}
		heap.Init(h)
		for j, s := range row {
			if i == j {
				continue
			}
			if h.Len() < k {
				heap.Push(h, types.Neighbor{Item: j, Sim: s})
			} else if s > (*h)[0].Sim {
				heap.Pop(h)
				heap.Push(h, types.Neighbor{Item: j, Sim: s})
			}
		}
		n := h.Len()
		buf := make([]types.Neighbor, n)
		for idx := 0; idx < n; idx++ {
			buf[idx] = heap.Pop(h).(types.Neighbor)
		}

		// invertimos -> mayor a menor
		for l, r := 0, n-1; l < r; l, r = l+1, r-1 {
			buf[l], buf[r] = buf[r], buf[l]
		}
		out[i] = buf
	}
	return out
}
