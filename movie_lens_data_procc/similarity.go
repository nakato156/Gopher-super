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
)

const (
	dataDir        = "../dataset/ml-latest-small"
	ratingsFile    = "ratings.csv"
	moviesFile     = "movies.csv"
	topKNeighborsN = 20  // vecinos por ítem
	minCoRatings   = 2   // mínimo de usuarios en común para considerar similitud
	userBatchSize  = 200 // tamaño de lote por worker
)

// ---------- Tipos básicos ----------

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

// ---------- Carga de datos ----------

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

	// saltar encabezado
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

// ---------- Estructuras derivadas ----------

// userRatings[u][i] = rating
// itemUsers[i][u] = rating
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

// ---------- Construcción concurrente de similitudes ----------

type partialAcc struct {
	dot  map[int]map[int]float64 // dot[i][j] con i<j
	co   map[int]map[int]int     // co-ratings
	norm map[int]float64         // ||i||^2
}

func newPartialAcc() *partialAcc {
	return &partialAcc{
		dot:  make(map[int]map[int]float64),
		co:   make(map[int]map[int]int),
		norm: make(map[int]float64),
	}
}

func (p *partialAcc) addDot(i, j int, v float64) {
	if i > j {
		i, j = j, i
	}
	m, ok := p.dot[i]
	if !ok {
		m = make(map[int]float64)
		p.dot[i] = m
	}
	m[j] += v
	cm, ok := p.co[i]
	if !ok {
		cm = make(map[int]int)
		p.co[i] = cm
	}
	cm[j]++
}

func (p *partialAcc) addNorm(i int, v float64) {
	p.norm[i] += v
}

func mergeAcc(dst, src *partialAcc) {
	for i, row := range src.dot {
		dr, ok := dst.dot[i]
		if !ok {
			dr = make(map[int]float64, len(row))
			dst.dot[i] = dr
		}
		for j, v := range row {
			dr[j] += v
		}
	}
	for i, row := range src.co {
		dr, ok := dst.co[i]
		if !ok {
			dr = make(map[int]int, len(row))
			dst.co[i] = dr
		}
		for j, c := range row {
			dr[j] += c
		}
	}
	for i, v := range src.norm {
		dst.norm[i] += v
	}
}

// worker: procesa un bloque de usuarios y acumula productos punto & normas
func processUserBlock(users []int, userRatings map[int]map[int]float64) *partialAcc {
	acc := newPartialAcc()
	for _, u := range users {
		ru := userRatings[u]
		if len(ru) == 0 {
			continue
		}

		// normas por ítem (sum r^2)
		for i, r := range ru {
			acc.addNorm(i, r*r)
		}

		// pares de ítems del mismo usuario
		// convertir a slice para índice estable
		items := make([]int, 0, len(ru))
		for i := range ru {
			items = append(items, i)
		}
		sort.Ints(items)
		for a := 0; a < len(items); a++ {
			i := items[a]
			ri := ru[i]
			for b := a + 1; b < len(items); b++ {
				j := items[b]
				rj := ru[j]
				acc.addDot(i, j, ri*rj)
			}
		}
	}
	return acc
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

func buildSimilaritiesConcurrent(userRatings map[int]map[int]float64) (SimMatrix, map[int]float64) {
	users := make([]int, 0, len(userRatings))
	for u := range userRatings {
		users = append(users, u)
	}
	sort.Ints(users)

	blocks := chunkInts(users, userBatchSize)

	// Fan-out workers -> producen parciales
	// combiner único fusiona
	workerCount := runtime.NumCPU()
	workCh := make(chan []int)
	partCh := make(chan *partialAcc, workerCount*2)
	var wg sync.WaitGroup

	combinerDone := make(chan struct{})
	global := newPartialAcc()

	// combiner
	go func() {
		for p := range partCh {
			mergeAcc(global, p)
		}
		close(combinerDone)
	}()

	// workers
	worker := func() {
		defer wg.Done()
		for blk := range workCh {
			p := processUserBlock(blk, userRatings)
			partCh <- p
		}
	}
	wg.Add(workerCount)
	for w := 0; w < workerCount; w++ {
		go worker()
	}

	// alimentar trabajos
	for _, blk := range blocks {
		workCh <- blk
	}
	close(workCh)

	wg.Wait()
	close(partCh)
	<-combinerDone

	// convertir a similitudes coseno con filtro por co-ratings
	sim := make(SimMatrix, len(global.dot))
	for i, row := range global.dot {
		ni := math.Sqrt(global.norm[i])
		if ni == 0 {
			continue
		}
		for j, dot := range row {
			if global.co[i][j] < minCoRatings {
				continue
			}
			nj := math.Sqrt(global.norm[j])
			if nj == 0 {
				continue
			}
			val := dot / (ni * nj)
			if val == 0 {
				continue
			}
			if sim[i] == nil {
				sim[i] = make(map[int]float64)
			}
			if sim[j] == nil {
				sim[j] = make(map[int]float64)
			}
			sim[i][j] = val
			sim[j][i] = val
		}
	}

	// devolvemos normas
	norms := make(map[int]float64, len(global.norm))
	for i, s := range global.norm {
		norms[i] = math.Sqrt(s)
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
			} else if k > 0 && s > (*h)[0].Sim {
				heap.Pop(h)
				heap.Push(h, Neighbor{Item: j, Sim: s})
			}
		}
		// volcar en orden descendente
		n := h.Len()
		buf := make([]Neighbor, n)
		for idx := n - 1; idx >= 0; idx-- {
			buf[idx] = heap.Pop(h).(Neighbor)
		}

		// se inverte para descendente
		for l, r := 0, len(buf)-1; l < r; l, r = l+1, r-1 {
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

func main() {
	fmt.Println("GoFlix · Item-based CF (cosine) · concurrente en un solo nodo")

	// 1) Cargar datos
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

	// Similitudes concurrentes
	fmt.Println("Calculando similitudes coseno (concurrente)…")
	sim, _ := buildSimilaritiesConcurrent(userRatings)
	fmt.Printf("Ítems con vecindad calculada: %d\n", len(sim))

	// Top-K vecinos por ítem
	fmt.Printf("Seleccionando top-%d vecinos por ítem…\n", topKNeighborsN)
	nbrs := topKNeighbors(sim, topKNeighborsN)

	// elegir el primer usuario y recomendar
	if len(users) == 0 {
		fmt.Println("No hay usuarios.")
		return
	}
	demoUser := users[0]
	fmt.Printf("Ejemplo de recomendaciones para el usuario %d\n", demoUser)
	recs := recommendTopN(demoUser, 10, userRatings, nbrs, titles)
	if len(recs) == 0 {
		fmt.Println("No se pudieron generar recomendaciones para el usuario de ejemplo.")
		return
	}
	fmt.Println(humanList(recs))

	// predecir rating de una película no vista (si existe)
	// se toma la primera recomendación como candidata.
	if len(recs) > 0 {
		item := recs[0].MovieID
		if p, ok := predictForUserItem(demoUser, item, userRatings, nbrs); ok {
			fmt.Printf("Predicción para user %d sobre movie %d (%s): %.3f\n",
				demoUser, item, titles[item], p)
		}
	}
}
