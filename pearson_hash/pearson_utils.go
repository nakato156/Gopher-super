package main

// Generar tabla de permutación de Pearson
func generatePearsonTable() [256]byte {
	var table [256]byte
	for i := 0; i < 256; i++ {
		table[i] = byte((i*7 + 13) % 256)
	}
	return table
}

var pearsonTable = generatePearsonTable()

// Algoritmo de Pearson secuencial
func PearsonSequential(input string) byte {
	hash := byte(0)
	for _, char := range input {
		hash = pearsonTable[hash^byte(char)]
	}
	return hash
}

// Algoritmo de Pearson con canales
func PearsonWithChannels(inputs []string) []byte {
	results := make([]byte, len(inputs))

	// Canal para recibir resultados
	resultChan := make(chan struct {
		index int
		hash  byte
	}, len(inputs))

	// Procesar cada cadena en una goroutine
	for i, input := range inputs {
		go func(index int, str string) {
			hash := PearsonSequential(str)
			resultChan <- struct {
				index int
				hash  byte
			}{index, hash}
		}(i, input)
	}

	// Recopilar resultados
	for i := 0; i < len(inputs); i++ {
		result := <-resultChan
		results[result.index] = result.hash
	}

	return results
}
