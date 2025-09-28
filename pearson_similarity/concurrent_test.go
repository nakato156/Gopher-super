package main

import "fmt"

// Función de prueba para el algoritmo concurrente de correlación
func testConcurrent() {
	// Crear vectores de prueba
	testVectors := createTestVectors()

	fmt.Println("=== Algoritmo de Correlación de Pearson Concurrente ===")
	fmt.Println("Procesando múltiples pares de vectores en paralelo...")
	fmt.Println()

	// Procesar todos los pares de vectores concurrentemente
	results := PearsonCorrelationWithChannels(testVectors)

	fmt.Println("Resultados:")
	fmt.Println("Par\t\tVector X\t\tVector Y\t\tCorrelación")
	fmt.Println("----------------------------------------------------------------")

	for i, pair := range testVectors {
		mid := len(pair) / 2
		x := pair[:mid]
		y := pair[mid:]
		correlation := results[i]

		fmt.Printf("Par %d:\n", i+1)
		fmt.Printf("  X: %v\n", x)
		fmt.Printf("  Y: %v\n", y)
		fmt.Printf("  Correlación: %.4f\n", correlation)

		// Interpretar el resultado
		if correlation > 0.7 {
			fmt.Printf("  Interpretación: Correlación fuerte positiva\n")
		} else if correlation < -0.7 {
			fmt.Printf("  Interpretación: Correlación fuerte negativa\n")
		} else if correlation > 0.3 {
			fmt.Printf("  Interpretación: Correlación moderada positiva\n")
		} else if correlation < -0.3 {
			fmt.Printf("  Interpretación: Correlación moderada negativa\n")
		} else {
			fmt.Printf("  Interpretación: Correlación débil o nula\n")
		}
		fmt.Println()
	}
}

// Función de prueba para el algoritmo concurrente de correlación
