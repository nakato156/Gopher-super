package main

import (
	"math"
)

// Calcular la media de un vector
func calculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

// Calcular la desviación estándar de un vector
func calculateStandardDeviation(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range data {
		sum += math.Pow(value-mean, 2)
	}
	return math.Sqrt(sum / float64(len(data)))
}

// Algoritmo de correlación de Pearson secuencial
func PearsonCorrelationSequential(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	meanX := calculateMean(x)
	meanY := calculateMean(y)

	numerator := 0.0
	sumXSquared := 0.0
	sumYSquared := 0.0

	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumXSquared += dx * dx
		sumYSquared += dy * dy
	}

	if sumXSquared == 0 || sumYSquared == 0 {
		return 0
	}

	return numerator / math.Sqrt(sumXSquared*sumYSquared)
}

// Algoritmo de correlación de Pearson con canales
func PearsonCorrelationWithChannels(pairs [][]float64) []float64 {
	results := make([]float64, len(pairs))

	// Canal para recibir resultados
	resultChan := make(chan struct {
		index int
		corr  float64
	}, len(pairs))

	// Procesar cada par de vectores en una goroutine
	for i, pair := range pairs {
		go func(index int, vectors []float64) {
			// Dividir el vector en dos mitades (x e y)
			mid := len(vectors) / 2
			if mid == 0 {
				resultChan <- struct {
					index int
					corr  float64
				}{index, 0}
				return
			}

			x := vectors[:mid]
			y := vectors[mid:]

			corr := PearsonCorrelationSequential(x, y)
			resultChan <- struct {
				index int
				corr  float64
			}{index, corr}
		}(i, pair)
	}

	// Recopilar resultados
	for i := 0; i < len(pairs); i++ {
		result := <-resultChan
		results[result.index] = result.corr
	}

	return results
}

// Función auxiliar para crear pares de vectores de prueba
func createTestVectors() [][]float64 {
	return [][]float64{
		// Correlación fuerte positiva (~0.8)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		// Correlación moderada positiva (~0.6)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 3, 2, 5, 4, 7, 6, 9, 8, 11},
		// Correlación débil positiva (~0.3)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 2, 1, 4, 3, 6, 5, 8, 7, 10, 9},
		// Correlación débil negativa (~-0.3)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		// Correlación moderada negativa (~-0.6)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		// Correlación fuerte negativa (~-0.8)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2},
		// Sin correlación (~0.0)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
		// Correlación muy débil (~0.1)
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 3, 2, 4, 3, 5, 4, 6, 5, 7, 6},
	}
}
