package main

import (
	"algsim/similarity"
	"runtime"
)

func main() {
	var rows int = 10_000
	var rowSize int = 2048
	var batch int = 8 * 4
	var nRequests int = 20
	var workers int = 3 * runtime.NumCPU()

	// JACARD
	// Construye matriz base M
	M := similarity.BuildMatrix(rows, rowSize, 2000)
	// Vector A aleatorio
	A := *similarity.BuildAVector(rowSize)

	// Test secuencial
	similarity.TestJaccardSeq(A, M)

	// Test concurrente
	similarity.TestJaccardCon(M, A, batch, workers, nRequests)

	// COSENO
	// Construye matriz base Mfloat
	Mfloat := similarity.CreateRandomFPMatrix(rows, rowSize)

	// Test secuencial
	similarity.TestCosine(Mfloat, rows, rowSize)

	// Test concurrente
	similarity.TestCosineCon(Mfloat, rows, rowSize)

	// PEARSON
}
