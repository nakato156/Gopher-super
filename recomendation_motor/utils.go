package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

// fileExists verifica si un archivo existe
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// loadGob carga una estructura desde un archivo .gob
func loadGob(filepath string, data interface{}) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error abriendo archivo %s: %w", filepath, err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("error decodificando datos de %s: %w", filepath, err)
	}
	return nil
}

// ============================================================================
// FUNCIONES DE MÉTRICAS DE RENDIMIENTO
// ============================================================================

// StartMetrics inicia el cronómetro de métricas
func StartMetrics() PerformanceMetrics {
	return PerformanceMetrics{
		StartTime: time.Now().UnixNano(),
	}
}

// EndMetrics finaliza el cronómetro y calcula métricas
func EndMetrics(metrics PerformanceMetrics, itemsProcessed int64, workers int) PerformanceMetrics {
	metrics.EndTime = time.Now().UnixNano()
	metrics.Duration = metrics.EndTime - metrics.StartTime
	metrics.DurationMs = float64(metrics.Duration) / 1e6
	metrics.DurationSec = float64(metrics.Duration) / 1e9
	metrics.ItemsProcessed = itemsProcessed
	metrics.Workers = workers

	if metrics.DurationSec > 0 {
		metrics.ItemsPerSec = float64(itemsProcessed) / metrics.DurationSec
		metrics.ItemsPerMs = float64(itemsProcessed) / metrics.DurationMs
	}

	return metrics
}

// CalculateSpeedup calcula el speedup vs tiempo secuencial
func CalculateSpeedup(sequentialTime, parallelTime int64) float64 {
	if parallelTime == 0 {
		return 0
	}
	return float64(sequentialTime) / float64(parallelTime)
}

// CalculateEfficiency calcula la eficiencia del paralelismo
func CalculateEfficiency(speedup float64, workers int) float64 {
	if workers == 0 {
		return 0
	}
	return speedup / float64(workers)
}

// CalculateScalability calcula la escalabilidad
func CalculateScalability(speedup float64, workers int) float64 {
	if workers <= 1 {
		return 1.0
	}
	return speedup / float64(workers-1)
}

// PrintMetrics imprime las métricas de rendimiento
func PrintMetrics(metrics PerformanceMetrics, title string) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    %-40s ║\n", title)
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")

	fmt.Printf("⏱️  TIEMPO:\n")
	fmt.Printf("   - Duración Total: %.2f ms (%.3f segundos)\n", metrics.DurationMs, metrics.DurationSec)
	fmt.Printf("   - Workers Utilizados: %d\n", metrics.Workers)

	fmt.Printf("\n📊 RENDIMIENTO:\n")
	fmt.Printf("   - Elementos Procesados: %d\n", metrics.ItemsProcessed)
	fmt.Printf("   - Elementos/segundo: %.2f\n", metrics.ItemsPerSec)
	fmt.Printf("   - Elementos/milisegundo: %.2f\n", metrics.ItemsPerMs)

	if metrics.Speedup > 0 {
		fmt.Printf("\n🚀 PARALELISMO:\n")
		fmt.Printf("   - Speedup: %.2fx\n", metrics.Speedup)
		fmt.Printf("   - Eficiencia: %.2f%%\n", metrics.Efficiency*100)
		if metrics.Scalability > 0 {
			fmt.Printf("   - Escalabilidad: %.2f\n", metrics.Scalability)
		}
	}

	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")
}

// BenchmarkSequential ejecuta una función de forma secuencial para benchmark
func BenchmarkSequential(fn func() error) (int64, error) {
	start := time.Now().UnixNano()
	err := fn()
	end := time.Now().UnixNano()
	return end - start, err
}

// BenchmarkParallel ejecuta una función de forma paralela para benchmark
func BenchmarkParallel(fn func() error, workers int) (int64, error) {
	start := time.Now().UnixNano()
	err := fn()
	end := time.Now().UnixNano()
	return end - start, err
}
