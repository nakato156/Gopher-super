package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go [secuencial|concurrente|benchmark]")
		fmt.Println("Ejemplos:")
		fmt.Println("  go run main.go secuencial")
		fmt.Println("  go run main.go concurrente")
		fmt.Println("  go run main.go benchmark")
		return
	}

	command := os.Args[1]

	switch command {
	case "secuencial":
		testSequential()
	case "concurrente":
		testConcurrent()
	case "benchmark":
		benchmarkAlgorithms()
	default:
		fmt.Printf("Comando desconocido: %s\n", command)
		fmt.Println("Comandos disponibles: secuencial, concurrente, benchmark")
	}
}

// Función de prueba para el algoritmo secuencial de correlación
func testSequential() {
	// Crear vectores de prueba
	testVectors := createTestVectors()

	// Solo mostrar los índices de correlación

	for _, pair := range testVectors {
		mid := len(pair) / 2
		x := pair[:mid]
		y := pair[mid:]

		correlation := PearsonCorrelationSequential(x, y)

		fmt.Printf("%.4f\n", correlation)
	}
}

// Función de prueba para el algoritmo concurrente de correlación
func testConcurrent() {
	// Crear vectores de prueba
	testVectors := createTestVectors()

	// Procesar todos los pares de vectores concurrentemente
	results := PearsonCorrelationWithChannels(testVectors)

	// Solo mostrar los índices de correlación
	for _, correlation := range results {
		fmt.Printf("%.4f\n", correlation)
	}
}
