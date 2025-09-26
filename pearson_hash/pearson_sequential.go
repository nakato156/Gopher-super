package main

import "fmt"

// Función de prueba para el algoritmo secuencial
func testSequential() {
	testCases := []string{"", "a", "hello", "world", "golang", "pearson"}

	fmt.Println("=== Algoritmo de Pearson Secuencial ===")
	fmt.Println("Entrada\t\tHash")
	fmt.Println("------------------------")

	for _, test := range testCases {
		hash := PearsonSequential(test)
		fmt.Printf("%-10s\t0x%02X\n", test, hash)
	}
}

func main() {
	testSequential()
}
