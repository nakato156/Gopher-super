package main

import (
	"bufio"
	"container/heap"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dataDir        = "../dataset/ml-latest-small"
	ratingsFile    = "ratings.csv"
	moviesFile     = "movies.csv"
	topKNeighborsN = 20  // vecinos por ítem
	minCoRatings   = 2   // mínimo de usuarios en común para considerar similitud
	userBatchSize  = 200 // tamaño de lote por worker
)

type Rating struct {
	UserID  int
	MovieID int
	Rating  float64
}

type Neighbor struct {
	Item int
	Sim  float64
}

type NeighborHeap []Neighbor

func (h NeighborHeap) Len() int            { return len(h) }
func (h NeighborHeap) Less(i, j int) bool  { return h[i].Sim < h[j].Sim } // min-heap
func (h NeighborHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *NeighborHeap) Push(x interface{}) { *h = append(*h, x.(Neighbor)) }
func (h *NeighborHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

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

func mustOpen(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}

func loadRatings(path string) ([]Rating, error) {
	f := mustOpen(path)
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1

	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("leyendo encabezado ratings: %w", err)
	}

	var ratings []Rating
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("leyendo ratings: %w", err)
		}
		if len(rec) < 3 {
			continue
		}
		uid, _ := strconv.Atoi(rec[0])
		mid, _ := strconv.Atoi(rec[1])
		val, _ := strconv.ParseFloat(rec[2], 64)
		ratings = append(ratings, Rating{UserID: uid, MovieID: mid, Rating: val})
	}
	return ratings, nil
}

func loadMovieTitles(path string) (map[int]string, error) {
	f := mustOpen(path)
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("leyendo encabezado movies: %w", err)
	}

	titles := make(map[int]string, 12000)
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("leyendo movies: %w", err)
		}
		if len(rec) < 3 {
			continue
		}
		mid, _ := strconv.Atoi(rec[0])
		title := rec[1]
		titles[mid] = title
	}
	return titles, nil
}

func buildIndexes(ratings []Rating) (map[int]map[int]float64, map[int]map[int]float64, []int, []int) {
	userRatings := make(map[int]map[int]float64, 1000)
	itemUsers := make(map[int]map[int]float64, 11000)

	userSet := make(map[int]struct{})
	itemSet := make(map[int]struct{})

	for _, rt := range ratings {
		if _, ok := userRatings[rt.UserID]; !ok {
			userRatings[rt.UserID] = make(map[int]float64)
		}
		userRatings[rt.UserID][rt.MovieID] = rt.Rating

		if _, ok := itemUsers[rt.MovieID]; !ok {
			itemUsers[rt.MovieID] = make(map[int]float64)
		}
		itemUsers[rt.MovieID][rt.UserID] = rt.Rating

		userSet[rt.UserID] = struct{}{}
		itemSet[rt.MovieID] = struct{}{}
	}

	users := make([]int, 0, len(userSet))
	for u := range userSet {
		users = append(users, u)
	}
	items := make([]int, 0, len(itemSet))
	for i := range itemSet {
		items = append(items, i)
	}
	sort.Ints(users)
	sort.Ints(items)
	return userRatings, itemUsers, users, items
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

type SimRow map[int]float64          // vecinos j -> sim(i,j)
type SimMatrix map[int]SimRow        // i -> (j->sim)
type NeighborList map[int][]Neighbor // i -> topK vecinos

func buildSimilaritiesConcurrent(userRatings map[int]map[int]float64, workerCount int) (SimMatrix, map[int]float64) {
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
		if co < minCoRatings {
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
func topKNeighbors(sim SimMatrix, k int) NeighborList {
	out := make(NeighborList, len(sim))
	for i, row := range sim {
		h := &NeighborHeap{}
		heap.Init(h)
		for j, s := range row {
			if i == j {
				continue
			}
			if h.Len() < k {
				heap.Push(h, Neighbor{Item: j, Sim: s})
			} else if s > (*h)[0].Sim {
				heap.Pop(h)
				heap.Push(h, Neighbor{Item: j, Sim: s})
			}
		}
		n := h.Len()
		buf := make([]Neighbor, n)
		for idx := 0; idx < n; idx++ {
			buf[idx] = heap.Pop(h).(Neighbor)
		}

		// invertimos -> mayor a menor
		for l, r := 0, n-1; l < r; l, r = l+1, r-1 {
			buf[l], buf[r] = buf[r], buf[l]
		}
		out[i] = buf
	}
	return out
}

func predictForUserItem(u int, i int, userRatings map[int]map[int]float64, nbrs NeighborList) (float64, bool) {
	ru, ok := userRatings[u]
	if !ok {
		return 0, false
	}
	neighbors, ok := nbrs[i]
	if !ok || len(neighbors) == 0 {
		return 0, false
	}
	num := 0.0
	den := 0.0
	for _, nb := range neighbors {
		if r, ok := ru[nb.Item]; ok {
			num += nb.Sim * r
			den += math.Abs(nb.Sim)
		}
	}
	if den == 0 {
		return 0, false
	}
	return num / den, true
}

type Rec struct {
	MovieID int
	Score   float64
	Title   string
}

func recommendTopN(u int, N int, userRatings map[int]map[int]float64, nbrs NeighborList, titles map[int]string) []Rec {
	ru := userRatings[u]
	seen := make(map[int]struct{}, len(ru))
	for mid := range ru {
		seen[mid] = struct{}{}
	}

	candidates := make(map[int]struct{})
	// candidatos: vecinos de lo que ya vio
	for mid := range ru {
		for _, nb := range nbrs[mid] {
			if _, ok := seen[nb.Item]; !ok {
				candidates[nb.Item] = struct{}{}
			}
		}
	}

	type scored struct {
		id  int
		val float64
	}
	var scores []scored
	for i := range candidates {
		if p, ok := predictForUserItem(u, i, userRatings, nbrs); ok {
			scores = append(scores, scored{id: i, val: p})
		}
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].val > scores[b].val })

	if len(scores) > N {
		scores = scores[:N]
	}
	recs := make([]Rec, 0, len(scores))
	for _, s := range scores {
		recs = append(recs, Rec{MovieID: s.id, Score: s.val, Title: titles[s.id]})
	}
	return recs
}

func humanList(recs []Rec) string {
	var b strings.Builder
	for i, r := range recs {
		fmt.Fprintf(&b, "%2d. (%.3f) %s [id=%d]\n", i+1, r.Score, safeTitle(r.Title), r.MovieID)
	}
	return b.String()
}

func safeTitle(s string) string {
	if s == "" {
		return "<sin título>"
	}
	return s
}

type benchRow struct {
	Workers int
	Millis  int64
	Speedup float64
}

func benchmarkWorkers(userRatings map[int]map[int]float64) ([]benchRow, int) {
	// GOMAXPROCS = NumCPU
	// para que > NumCPU workers muestren overhead
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	maxWorkers := 4 * cpus

	// ejecutamos con workers de 1 - 2*CPU
	results := make([]benchRow, 0, maxWorkers)

	// baseline con 1 worker
	start := time.Now()
	_, _ = buildSimilaritiesConcurrent(userRatings, 1)
	baseMs := time.Since(start).Milliseconds()
	if baseMs == 0 {
		baseMs = 1
	}
	results = append(results, benchRow{Workers: 1, Millis: baseMs, Speedup: 1.0})

	// resto
	bestIdx := 0
	bestMs := baseMs
	for w := 2; w <= maxWorkers; w++ {
		t0 := time.Now()
		_, _ = buildSimilaritiesConcurrent(userRatings, w)
		ms := time.Since(t0).Milliseconds()
		sp := float64(baseMs) / float64(ms)
		results = append(results, benchRow{Workers: w, Millis: ms, Speedup: sp})
		if ms < bestMs {
			bestMs = ms
			bestIdx = len(results) - 1
		}
	}
	return results, results[bestIdx].Workers
}

func printBench(results []benchRow) {
	fmt.Println("\n=== Benchmark de goroutines para cálculo de similitudes ===")
	fmt.Printf("GOMAXPROCS = %d (NumCPU)\n", runtime.NumCPU())
	fmt.Printf("%8s  %12s  %8s\n", "workers", "ms", "speedup")
	for _, r := range results {
		fmt.Printf("%8d  %12d  %8.2f\n", r.Workers, r.Millis, r.Speedup)
	}
}

func main() {
	fmt.Println("GoFlix · Item-based CF (cosine) · concurrente en un solo nodo")

	rPath := filepath.Join(dataDir, ratingsFile)
	mPath := filepath.Join(dataDir, moviesFile)

	fmt.Println("Cargando ratings desde:", rPath)
	ratings, err := loadRatings(rPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("Cargando títulos desde:", mPath)
	titles, err := loadMovieTitles(mPath)
	if err != nil {
		panic(err)
	}

	// Índices (user->items, item->users)
	fmt.Println("Construyendo índices…")
	userRatings, _, users, _ := buildIndexes(ratings)
	fmt.Printf("Usuarios: %d, Ratings: %d\n", len(users), len(ratings))

	// Benchmark de paralelismo (solo similitudes)
	results, bestWorkers := benchmarkWorkers(userRatings)
	printBench(results)
	fmt.Printf("\nMejor configuración observada: %d workers\n", bestWorkers)

	// Construir similitudes y vecindades con la mejor configuración
	fmt.Println("\nReconstruyendo similitudes con la mejor configuración…")
	sim, _ := buildSimilaritiesConcurrent(userRatings, bestWorkers)
	fmt.Printf("Ítems con vecindad calculada: %d\n", len(sim))

	fmt.Printf("Seleccionando top-%d vecinos por ítem…\n", topKNeighborsN)
	nbrs := topKNeighbors(sim, topKNeighborsN)

	// recomendar a un usuario
	if len(users) == 0 {
		fmt.Println("No hay usuarios.")
		return
	}
	demoUser := users[0]
	fmt.Printf("\nEjemplo de recomendaciones para el usuario %d\n", demoUser)
	recs := recommendTopN(demoUser, 10, userRatings, nbrs, titles)
	if len(recs) == 0 {
		fmt.Println("No se pudieron generar recomendaciones para el usuario de ejemplo.")
		return
	}
	fmt.Println(humanList(recs))

	// Predicción de una película candidata
	if len(recs) > 0 {
		item := recs[0].MovieID
		if p, ok := predictForUserItem(demoUser, item, userRatings, nbrs); ok {
			fmt.Printf("Predicción para user %d sobre movie %d (%s): %.3f\n",
				demoUser, item, titles[item], p)
		}
	}
}
