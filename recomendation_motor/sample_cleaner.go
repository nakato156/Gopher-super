package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// getRequiredColumns define las columnas que se deben mantener
func getRequiredColumns() []string {
	return []string{
		// Variables Continuas (9):
		"author.playtime_forever",
		"author.playtime_last_two_weeks",
		"author.playtime_at_review",
		"votes_helpful",
		"votes_funny",
		"weighted_vote_score",
		"timestamp_created",
		"timestamp_updated",
		"author.last_played",
		// Variables Discretas/Conteo (5):
		"author.num_games_owned",
		"author.num_reviews",
		"comment_count",
		"app_id",
		"review_id",
		// Variables Binarias (4):
		"recommended",
		"steam_purchase",
		"received_for_free",
		"written_during_early_access",
		// Identificadores (2):
		"author.steamid",
		"app_name", // Nombre del juego
	}
}

// cleanAndSampleCSV limpia el CSV y toma una muestra del 10%
func cleanAndSampleCSV(inputFile, outputFile string, sampleRate float64) error {
	// Abrir archivo de entrada
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error abriendo archivo de entrada: %w", err)
	}
	defer inFile.Close()

	// Crear archivo de salida
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creando archivo de salida: %w", err)
	}
	defer outFile.Close()

	// Crear writer CSV
	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Leer archivo línea por línea con buffer grande
	reader := bufio.NewReaderSize(inFile, 1024*1024) // Buffer de 1MB

	// Leer header
	headerLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error leyendo header: %w", err)
	}

	// Parsear header
	headerReader := csv.NewReader(strings.NewReader(headerLine))
	headerRecord, err := headerReader.Read()
	if err != nil {
		return fmt.Errorf("error parseando header: %w", err)
	}

	// Encontrar índices de las columnas necesarias
	columnIndices := findColumnIndices(headerRecord, getRequiredColumns())
	if len(columnIndices) == 0 {
		return fmt.Errorf("no se encontraron columnas necesarias en el header")
	}

	// Escribir header limpio
	cleanHeader := make([]string, len(columnIndices))
	for i, idx := range columnIndices {
		cleanHeader[i] = headerRecord[idx]
	}

	if err := writer.Write(cleanHeader); err != nil {
		return fmt.Errorf("error escribiendo header: %w", err)
	}

	fmt.Printf("📋 Columnas encontradas: %d de %d necesarias\n", len(columnIndices), len(getRequiredColumns()))
	fmt.Printf("📝 Header limpio: %s\n", strings.Join(cleanHeader, ", "))

	// Inicializar generador de números aleatorios
	rand.Seed(time.Now().UnixNano())

	// Procesar filas
	rowsProcessed := 0
	rowsWritten := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Saltar líneas con errores
			continue
		}

		rowsProcessed++

		// Mostrar progreso cada 100,000 filas
		if rowsProcessed%100000 == 0 {
			fmt.Printf("📊 Procesadas: %d filas...\n", rowsProcessed)
		}

		// Aplicar muestreo aleatorio
		if rand.Float64() > sampleRate {
			continue
		}

		// Parsear línea CSV
		lineReader := csv.NewReader(strings.NewReader(line))
		lineReader.FieldsPerRecord = -1 // Permitir campos variables

		record, err := lineReader.Read()
		if err != nil {
			// Saltar filas con errores de parseo
			continue
		}

		// Verificar que tenemos suficientes columnas
		if len(record) < len(columnIndices) {
			continue
		}

		// Extraer solo las columnas necesarias
		cleanRecord := make([]string, len(columnIndices))
		validRow := true

		for i, idx := range columnIndices {
			if idx < len(record) {
				cleanRecord[i] = record[idx]
			} else {
				validRow = false
				break
			}
		}

		if !validRow {
			continue
		}

		// Escribir fila limpia
		if err := writer.Write(cleanRecord); err != nil {
			return fmt.Errorf("error escribiendo fila: %w", err)
		}

		rowsWritten++
	}

	fmt.Printf("\n📊 Procesamiento completado:\n")
	fmt.Printf("   - Filas procesadas: %d\n", rowsProcessed)
	fmt.Printf("   - Filas en muestra: %d (%.1f%%)\n", rowsWritten, float64(rowsWritten)/float64(rowsProcessed)*100)
	fmt.Printf("   - Archivo de salida: %s\n", outputFile)

	return nil
}

// findColumnIndices encuentra los índices de las columnas necesarias
func findColumnIndices(header []string, requiredColumns []string) []int {
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	var indices []int
	for _, reqCol := range requiredColumns {
		if idx, ok := columnMap[reqCol]; ok {
			indices = append(indices, idx)
		} else {
			fmt.Printf("⚠️  Columna requerida no encontrada: %s\n", reqCol)
		}
	}
	return indices
}

func runCleaner() {
	fmt.Printf("╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║   LIMPIADOR Y MUESTREADOR DE CSV                          ║\n")
	fmt.Printf("║   Generando muestra con columnas limpiadas                ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	// Cargar configuración
	configFile := "config.json"
	systemConfig, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Error cargando configuración: %v\n", err)
		fmt.Printf("🔧 Usando configuración por defecto\n")
		systemConfig = DefaultConfig()
	}

	// Mostrar configuración actual
	PrintConfig(systemConfig)

	inputFile := "../steam_reviews.csv"
	outputFile := fmt.Sprintf("steam_reviews_sample_%dpct.csv", systemConfig.Sampling.Percentage)
	sampleRate := float64(systemConfig.Sampling.Percentage) / 100.0

	// Configurar semilla aleatoria
	rand.Seed(int64(systemConfig.Sampling.RandomSeed))

	fmt.Printf("\n🔄 Procesando archivo: %s\n", inputFile)
	fmt.Printf("📊 Tasa de muestreo: %.1f%%\n", sampleRate*100)
	fmt.Printf("🎲 Random Seed: %d\n", systemConfig.Sampling.RandomSeed)
	fmt.Printf("📂 Archivo de salida: %s\n\n", outputFile)

	if err := cleanAndSampleCSV(inputFile, outputFile, sampleRate); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Proceso completado exitosamente!\n")
	fmt.Printf("📁 Archivo generado: %s\n", outputFile)
}
