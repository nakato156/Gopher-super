package cosine

import (
	"fmt"
	"math"
	"time"
)

func TestCosine(M [][]float64, rows int, rowSize int) {
	if len(M) == 0 || rows <= 0 || rowSize <= 0 {
		return
	}
	if rows > len(M) {
		rows = len(M)
	}

	cols := rowSize
	for i := 0; i < rows; i++ {
		if l := len(M[i]); l < cols {
			cols = l
		}
	}
	if cols == 0 {
		return
	}

	vectors := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		vectors[i] = M[i][:cols]
	}

	start := time.Now()

	invNorms := make([]float64, rows)
	for i := 0; i < rows; i++ {
		row := vectors[i]
		var sum float64
		j := 0
		for ; j+7 < cols; j += 8 {
			sum += row[j+0]*row[j+0] +
				row[j+1]*row[j+1] +
				row[j+2]*row[j+2] +
				row[j+3]*row[j+3] +
				row[j+4]*row[j+4] +
				row[j+5]*row[j+5] +
				row[j+6]*row[j+6] +
				row[j+7]*row[j+7]
		}
		for ; j < cols; j++ {
			v := row[j]
			sum += v * v
		}
		if sum > 0 {
			invNorms[i] = 1 / math.Sqrt(sum)
		}
	}

	results := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		ai := vectors[i]
		invA := invNorms[i]
		dst := make([]float64, rows)
		if invA != 0 {
			for j := 0; j < rows; j++ {
				invB := invNorms[j]
				if invB == 0 {
					continue
				}
				bj := vectors[j]
				dst[j] = dotUnrolled8(ai, bj) * invA * invB
			}
		}
		results[i] = dst
	}

	elapsed := time.Since(start)
	fmt.Printf("Cosine Sim. NxN completada: %d usuarios, %d dimensiones\n", rows, cols)
	fmt.Printf("Cosine Sim. Tiempo total secuencial: %v\n", elapsed)
}
