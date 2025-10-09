package pearson

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Función de benchmark comparativo para correlación de Pearson
func BenchmarkAlgorithms(rows, rowSize int) (time.Duration, time.Duration) {
	// Crear vectores de prueba más grandes para el benchmark
	testVectors := createLargeTestVectors(rows, rowSize)

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

	return seqTime, chanTime
}

// Función auxiliar para crear vectores de prueba más grandes
func createLargeTestVectors(rows, rowSize int) [][]float64 {
	if rows <= 0 || rowSize <= 0 {
		return nil
	}

	pairLength := rowSize * 2
	vectors := make([][]float64, rows)
	data := make([]float64, rows*pairLength)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range data {
		data[i] = rng.Float64()*2 - 1
	}

	for i := 0; i < rows; i++ {
		start := i * pairLength
		vectors[i] = data[start : start+pairLength]
	}

	return vectors
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
