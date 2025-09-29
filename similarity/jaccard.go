package similarity

import (
	"fmt"
	"time"
)

func TestJaccardSeq(A []uint32, B *Matrix) []float64 {
	start := time.Now()
	if B == nil {
		return nil
	}
	Auniq := append([]uint32(nil), A...)
	Auniq = uniqueSorted(Auniq)
	res := make([]float64, len(B.Rows))
	for i := range B.Rows {
		res[i] = jaccardSorted(Auniq, B.Rows[i].Set)
	}
	elapsed := time.Since(start)
	fmt.Printf("Jaccard secuencial: %v\n", elapsed)
	return res
}
