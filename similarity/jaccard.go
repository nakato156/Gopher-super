package similarity

import (
	"context"
	"runtime"
	"sync"
)

func jaccardSeq(A, B []uint32) float64 {
	setA := make(map[uint32]struct{})
	setB := make(map[uint32]struct{})

	for _, num := range A {
		setA[num] = struct{}{}
	}

	for _, num := range B {
		setB[num] = struct{}{}
	}

	interseccion := 0

	for number := range setA {
		if _, found := setB[number]; found {
			interseccion++
		}
	}

	union := len(setA) + len(setB) - interseccion

	if union == 0 {
		return 0.0
	}

	return float64(interseccion) / float64(union)
}

func jaccardSortedU32(a, b []uint32) float64 {
	i, j := 0, 0
	la, lb := len(a), len(b)
	if la == 0 && lb == 0 {
		return 1.0
	}
	inter, uni := 0, 0
	for i < la && j < lb {
		av, bv := a[i], b[j]
		if av == bv {
			inter++
			uni++
			i++
			j++
		} else if av < bv {
			uni++
			i++
		} else {
			uni++
			j++
		}
	}
	uni += (la - i) + (lb - j)
	// uni > 0 por el guard clause de arriba
	return float64(inter) / float64(uni)
}

// froma concurrente con pool de workers y canal de jobs ---
func JaccardCon(ctx context.Context, sets [][]uint32, pairs []Pair, workers int) []float64 {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	// Preasignamos resultados; cada job escribe su índice directamente (evita canal de resultados)
	results := make([]float64, len(pairs))

	type job struct {
		idx    int
		aIndex int
		bIndex int
	}
	jobs := make(chan job, workers*4)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case jb, ok := <-jobs:
					if !ok {
						return
					}
					results[jb.idx] = jaccardSortedU32(sets[jb.aIndex], sets[jb.bIndex])
				}
			}
		}()
	}

	// Feed de trabajos
	for i, p := range pairs {
		select {
		case <-ctx.Done():
			break
		case jobs <- job{idx: i, aIndex: p.A, bIndex: p.B}:
		}
	}
	close(jobs)
	wg.Wait()
	return results
}
