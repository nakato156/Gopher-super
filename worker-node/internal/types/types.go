package types

// From data
type Rating struct {
	UserID  int
	MovieID int
	Rating  float64
}

// From engine
const (
	TopKNeighborsN = 30  // vecinos por ítem
	MinCoRatings   = 10  // mínimo de usuarios en común para considerar similitud
	UserBatchSize  = 200 // tamaño de lote por worker
)

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

type SimRow map[int]float64
type SimMatrix map[int]SimRow
type NeighborList map[int][]Neighbor

// From recommend
type Rec struct {
	MovieID int
	Score   float64
	Title   string
}

// From benchmark
type BenchRow struct {
	Workers int
	Millis  int64
	Speedup float64
}
