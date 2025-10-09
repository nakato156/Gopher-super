package main

import (
	"algsim/similarity"
	"algsim/similarity/cosine"
	"algsim/similarity/jaccard"
	"algsim/similarity/pearson"
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
)

func main() {
	var rows int = 10_000
	var rowSize int = 2048
	var batch int = 8 * 4
	var nRequests int = 20
	var workers int = 3 * runtime.NumCPU()

	type report struct {
		name       string
		sequential time.Duration
		concurrent time.Duration
	}

	reports := make([]report, 0, 3)

	// JACARD
	// Construye matriz base M
	M := jaccard.BuildMatrix(rows, rowSize, 2000)
	// Vector A aleatorio
	A := *jaccard.BuildAVector(rowSize)

	// Test secuencial
	start := time.Now()
	jaccard.TestJaccardSeq(A, M)
	jaccardSeqTime := time.Since(start)

	// Test concurrente
	start = time.Now()
	jaccard.TestJaccardCon(M, A, batch, workers, nRequests)
	jaccardConTime := time.Since(start)

	reports = append(reports, report{name: "Jaccard", sequential: jaccardSeqTime, concurrent: jaccardConTime})

	// COSENO
	// Construye matriz base Mfloat
	Mfloat := similarity.CreateRandomFPMatrix(rows, rowSize)

	// Test secuencial
	start = time.Now()
	cosine.TestCosine(Mfloat, rows, rowSize)
	cosineSeqTime := time.Since(start)

	// Test concurrente
	start = time.Now()
	cosine.TestCosineCon(Mfloat, rows, rowSize)
	cosineConTime := time.Since(start)

	reports = append(reports, report{name: "Coseno", sequential: cosineSeqTime, concurrent: cosineConTime})

	// PEARSON
	pearsonSeqTime, pearsonConTime := pearson.BenchmarkAlgorithms(rows, rowSize)
	reports = append(reports, report{name: "Pearson", sequential: pearsonSeqTime, concurrent: pearsonConTime})

	// Tabla resumen final
	fmt.Println("=== Tabla de tiempos de ejecución ===")
	fmt.Printf("%-10s %-18s %-18s %-10s\n", "Algoritmo", "Secuencial", "Concurrente", "Speedup")
	fmt.Println(strings.Repeat("-", 62))
	for _, r := range reports {
		speedup := math.Inf(1)
		if r.concurrent > 0 {
			speedup = float64(r.sequential) / float64(r.concurrent)
		}
		fmt.Printf("%-10s %-18s %-18s %-10.2f\n", r.name, r.sequential, r.concurrent, speedup)
	}
}
