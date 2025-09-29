package pearson

import "fmt"

// Función de prueba para el algoritmo secuencial de correlación
func TestSequential() {
	// Crear vectores de prueba
	testVectors := CreateTestVectors()

	fmt.Println("=== Algoritmo de Correlación de Pearson Secuencial ===")
	fmt.Println("Vector X\t\tVector Y\t\tCorrelación")
	fmt.Println("--------------------------------------------------------")

	for i, pair := range testVectors {
		mid := len(pair) / 2
		x := pair[:mid]
		y := pair[mid:]

		correlation := PearsonCorrelationSequential(x, y)

		fmt.Printf("Vector %d:\n", i+1)
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

// Función de prueba para el algoritmo secuencial de correlación
