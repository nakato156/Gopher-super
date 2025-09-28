package main

import (
	"fmt"
	"math"
	"time"
)

// Función de benchmark comparativo para correlación de Pearson
func benchmarkAlgorithms() {
	// Crear vectores de prueba más grandes para el benchmark
	testVectors := createLargeTestVectors()

	fmt.Println("=== Benchmark de Algoritmos de Correlación de Pearson ===")
	fmt.Printf("Procesando %d pares de vectores...\n", len(testVectors))
	fmt.Println()

	// Benchmark secuencial
	start := time.Now()
	seqResults := make([]float64, len(testVectors))
	for i, pair := range testVectors {
		mid := len(pair) / 2
		x := pair[:mid]
		y := pair[mid:]
		seqResults[i] = PearsonCorrelationSequential(x, y)
	}
	seqTime := time.Since(start)

	// Benchmark con canales
	start = time.Now()
	chanResults := PearsonCorrelationWithChannels(testVectors)
	chanTime := time.Since(start)

	// Mostrar resultados
	fmt.Println("Resultados del Benchmark:")
	fmt.Printf("Secuencial: %v\n", seqTime)
	fmt.Printf("Con Canales: %v\n", chanTime)
	fmt.Println()

	// Verificar que los resultados sean iguales (con tolerancia para float64)
	allEqual := true
	tolerance := 1e-10
	for i := range testVectors {
		if math.Abs(seqResults[i]-chanResults[i]) > tolerance {
			allEqual = false
			break
		}
	}

	if allEqual {
		fmt.Println("Ambos algoritmos producen resultados idénticos")
	} else {
		fmt.Println("Los algoritmos producen resultados diferentes")
	}

	// Mostrar diferencia de rendimiento
	if seqTime < chanTime {
		fmt.Printf("Secuencial es %.2fx más rápido\n", float64(chanTime)/float64(seqTime))
	} else {
		fmt.Printf("Con Canales es %.2fx más rápido\n", float64(seqTime)/float64(chanTime))
	}

	fmt.Println()
	fmt.Println("=== Comparación de Resultados ===")
	fmt.Println("Par\t\tSecuencial\tCon Canales\tDiferencia\t¿Iguales?")
	fmt.Println("----------------------------------------------------------------")

	for i := range testVectors {
		seq := seqResults[i]
		chanResult := chanResults[i]
		diff := math.Abs(seq - chanResult)
		equal := diff <= tolerance

		fmt.Printf("Par %d\t\t%.6f\t%.6f\t%.2e\t%t\n",
			i+1, seq, chanResult, diff, equal)
	}

	// Mostrar estadísticas adicionales
	fmt.Println()
	fmt.Println("=== Estadísticas de Correlación ===")
	fmt.Printf("Correlación promedio: %.4f\n", calculateMean(seqResults))
	fmt.Printf("Correlación máxima: %.4f\n", findMax(seqResults))
	fmt.Printf("Correlación mínima: %.4f\n", findMin(seqResults))
}

// Función auxiliar para crear vectores de prueba más grandes
func createLargeTestVectors() [][]float64 {
	return [][]float64{
		// Vectores con correlación perfecta positiva
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20},
		// Vectores con correlación perfecta negativa
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		// Vectores con correlación moderada
		{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20},
		// Vectores con correlación débil
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		// Vectores con correlación negativa moderada
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 18, 16, 14, 12, 10, 8, 6, 4, 2},
		// Vectores aleatorios
		{1, 4, 7, 2, 9, 3, 6, 8, 5, 10, 5, 2, 8, 1, 9, 4, 7, 3, 6, 10},
		{2, 5, 8, 3, 10, 4, 7, 9, 6, 11, 6, 3, 9, 2, 10, 5, 8, 4, 7, 11},
		{3, 6, 9, 4, 11, 5, 8, 10, 7, 12, 7, 4, 10, 3, 11, 6, 9, 5, 8, 12},
	}
}

// Funciones auxiliares para estadísticas
func findMax(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	max := data[0]
	for _, value := range data {
		if value > max {
			max = value
		}
	}
	return max
}

func findMin(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	min := data[0]
	for _, value := range data {
		if value < min {
			min = value
		}
	}
	return min
}

// Función de benchmark para algoritmos de correlación de Pearson
