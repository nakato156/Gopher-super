package main

import "fmt"

// Función de prueba para el algoritmo concurrente
func testConcurrent() {
	testCases := []string{"", "a", "hello", "world", "golang", "pearson"}

	fmt.Println("=== Algoritmo de Pearson Concurrente ===")
	fmt.Println("Entrada\t\tHash")
	fmt.Println("------------------------")

	// Procesar todas las cadenas concurrentemente
	results := PearsonWithChannels(testCases)

	for i, test := range testCases {
		fmt.Printf("%-10s\t0x%02X\n", test, results[i])
	}
}

func main() {
	testConcurrent()
}
