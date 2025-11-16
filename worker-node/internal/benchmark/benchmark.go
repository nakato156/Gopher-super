package benchmark

import (
	"fmt"
	"goflix/worker-node/internal/types"
	"runtime"
	"time"
)

type SimilarityBuilder func(userRatings map[int]map[int]float64, workerCount int) (types.SimMatrix, map[int]float64)

func BenchmarkWorkers(userRatings map[int]map[int]float64, builder SimilarityBuilder) ([]types.BenchRow, int) {
	// GOMAXPROCS = NumCPU
	// para que > NumCPU workers muestren overhead
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	maxWorkers := 4 * cpus

	// ejecutamos con workers de 1 - 2*CPU
	results := make([]types.BenchRow, 0, maxWorkers)

	// baseline con 1 worker
	start := time.Now()
	_, _ = builder(userRatings, 1)
	baseMs := time.Since(start).Milliseconds()
	if baseMs == 0 {
		baseMs = 1
	}
	results = append(results, types.BenchRow{Workers: 1, Millis: baseMs, Speedup: 1.0})

	// resto
	bestIdx := 0
	bestMs := baseMs
	for w := 2; w <= maxWorkers; w++ {
		t0 := time.Now()
		_, _ = builder(userRatings, w)
		ms := time.Since(t0).Milliseconds()
		sp := float64(baseMs) / float64(ms)
		results = append(results, types.BenchRow{Workers: w, Millis: ms, Speedup: sp})
		if ms < bestMs {
			bestMs = ms
			bestIdx = len(results) - 1
		}
	}
	return results, results[bestIdx].Workers
}

func PrintBench(results []types.BenchRow) {
	fmt.Println("=== Benchmark de goroutines para cálculo de similitudes ===")
	fmt.Printf("GOMAXPROCS = %d (NumCPU)", runtime.NumCPU())
	fmt.Printf("%8s  %12s  %8s", "workers", "ms", "speedup")
	for _, r := range results {
		fmt.Printf("%8d  %12d  %8.2f", r.Workers, r.Millis, r.Speedup)
	}
}
