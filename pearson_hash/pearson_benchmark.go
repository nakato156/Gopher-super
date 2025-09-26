package main

import (
	"fmt"
	"time"
)

// Función de benchmark comparativo
func benchmarkAlgorithms() {
	testStrings := []string{
		"Esta es una cadena muy larga para probar el rendimiento del algoritmo de Pearson en Go",
		"Otra cadena larga para probar la concurrencia en el algoritmo de Pearson",
		"Tercera cadena larga para comparar el rendimiento entre secuencial y concurrente",
		"Cuarta cadena larga para demostrar el uso de goroutines en Go",
		"Quinta cadena larga para probar el algoritmo de Pearson con concurrencia",
	}

	fmt.Println("=== Benchmark de Algoritmos de Pearson ===")
	fmt.Printf("Procesando %d cadenas...\n", len(testStrings))
	fmt.Println()

	// Benchmark secuencial
	start := time.Now()
	seqResults := make([]byte, len(testStrings))
	for i, str := range testStrings {
		seqResults[i] = PearsonSequential(str)
	}
	seqTime := time.Since(start)

	// Benchmark con canales
	start = time.Now()
	chanResults := PearsonWithChannels(testStrings)
	chanTime := time.Since(start)

	// Mostrar resultados
	fmt.Println("Resultados del Benchmark:")
	fmt.Printf("Secuencial: %v\n", seqTime)
	fmt.Printf("Con Canales: %v\n", chanTime)
	fmt.Println()

	// Verificar que los resultados sean iguales
	allEqual := true
	for i := range testStrings {
		if seqResults[i] != chanResults[i] {
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
	fmt.Println("Cadena\t\t\tSecuencial\tCon Canales\t¿Iguales?")
	fmt.Println("--------------------------------------------------------")

	for i, str := range testStrings {
		displayStr := str
		if len(displayStr) > 20 {
			displayStr = displayStr[:17] + "..."
		}

		seq := seqResults[i]
		chanResult := chanResults[i]
		equal := seq == chanResult

		fmt.Printf("%-20s\t0x%02X\t\t0x%02X\t\t%t\n",
			displayStr, seq, chanResult, equal)
	}
}

func main() {
	benchmarkAlgorithms()
}
