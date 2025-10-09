package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// parseRow parsea una fila del CSV de muestra
func parseRow(record []string) (*UserProfile, string, string, error) {
	if len(record) < 20 {
		return nil, "", "", fmt.Errorf("fila incompleta")
	}

	// Extraer campos relevantes
	steamID := record[18]    // author.steamid
	appID := record[12]      // app_id
	appName := record[19]    // app_name
	playtimeStr := record[0] // author.playtime_forever

	// Convertir playtime a float
	playtime, err := strconv.ParseFloat(playtimeStr, 64)
	if err != nil {
		return nil, "", "", fmt.Errorf("error parseando playtime: %w", err)
	}

	// Crear perfil de usuario
	profile := &UserProfile{
		UserID:   steamID,
		Games:    make(map[string]float64),
		Features: make([]float64, 18),
	}

	// Agregar juego al perfil
	profile.Games[appID] = playtime

	// Extraer características numéricas
	features := []string{
		record[0],  // author.playtime_forever
		record[1],  // author.playtime_last_two_weeks
		record[2],  // author.playtime_at_review
		record[3],  // votes_helpful
		record[4],  // votes_funny
		record[5],  // weighted_vote_score
		record[6],  // timestamp_created
		record[7],  // timestamp_updated
		record[8],  // author.last_played
		record[9],  // author.num_games_owned
		record[10], // author.num_reviews
		record[11], // comment_count
		record[13], // review_id
		record[14], // recommended
		record[15], // steam_purchase
		record[16], // received_for_free
		record[17], // written_during_early_access
	}

	// Convertir características a float64
	for i, featureStr := range features {
		if i < len(profile.Features) {
			if val, err := strconv.ParseFloat(featureStr, 64); err == nil {
				profile.Features[i] = val
			}
		}
	}

	return profile, appID, appName, nil
}

// ParseWorker procesa jobs de parsing
func ParseWorker(
	workerID int,
	jobs <-chan ParseJob,
	results chan<- ParseResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for job := range jobs {
		// Parsear línea CSV
		lineReader := csv.NewReader(strings.NewReader(job.Line))
		lineReader.FieldsPerRecord = -1

		record, err := lineReader.Read()
		if err != nil {
			results <- ParseResult{Valid: false}
			continue
		}

		profile, gameID, gameName, err := parseRow(record)
		if err != nil {
			results <- ParseResult{Valid: false}
			continue
		}

		results <- ParseResult{
			Profile:  profile,
			GameID:   gameID,
			GameName: gameName,
			Valid:    true,
		}
	}
}

// loadSampleCSV carga el CSV de muestra
func loadSampleCSV(filename string) ([]*UserProfile, map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	// Leer todas las líneas con buffer más grande
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Buffer de 1MB

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error leyendo archivo: %w", err)
	}

	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("archivo vacío o sin header")
	}

	// Parsear header
	headerReader := csv.NewReader(strings.NewReader(lines[0]))
	header, err := headerReader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error parseando header: %w", err)
	}

	fmt.Printf("📋 Header: %s\n", strings.Join(header, ", "))

	// Procesar filas
	profiles := make([]*UserProfile, 0)
	gameNames := make(map[string]string)
	userProfiles := make(map[string]*UserProfile)

	rowsProcessed := 0
	rowsValid := 0

	for i := 1; i < len(lines); i++ {
		rowsProcessed++

		if rowsProcessed%100000 == 0 {
			fmt.Printf("📊 Procesadas: %d filas...\n", rowsProcessed)
		}

		// Parsear línea CSV
		lineReader := csv.NewReader(strings.NewReader(lines[i]))
		lineReader.FieldsPerRecord = -1

		record, err := lineReader.Read()
		if err != nil {
			continue
		}

		profile, gameID, gameName, err := parseRow(record)
		if err != nil {
			continue
		}

		rowsValid++

		// Agregar o actualizar perfil de usuario
		if existingProfile, exists := userProfiles[profile.UserID]; exists {
			// Usuario ya existe, agregar juego
			existingProfile.Games[gameID] = profile.Games[gameID]
		} else {
			// Nuevo usuario
			userProfiles[profile.UserID] = profile
			profiles = append(profiles, profile)
		}

		// Agregar nombre real del juego
		gameNames[gameID] = gameName
	}

	fmt.Printf("📊 Filas procesadas: %d\n", rowsProcessed)
	fmt.Printf("📊 Filas válidas: %d\n", rowsValid)
	fmt.Printf("📊 Usuarios únicos: %d\n", len(profiles))
	fmt.Printf("📊 Juegos únicos: %d\n", len(gameNames))

	return profiles, gameNames, nil
}

// loadSampleCSVConcurrent carga el CSV de muestra usando concurrencia
func loadSampleCSVConcurrent(filename string, numWorkers int, bufferSize int) ([]*UserProfile, map[string]string, error) {
	// Iniciar métricas de rendimiento
	metrics := StartMetrics()

	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	// Leer todas las líneas con buffer más grande
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // Buffer de 1MB

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error leyendo archivo: %w", err)
	}

	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("archivo vacío o sin header")
	}

	// Parsear header
	headerReader := csv.NewReader(strings.NewReader(lines[0]))
	header, err := headerReader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error parseando header: %w", err)
	}

	fmt.Printf("📋 Header: %s\n", strings.Join(header, ", "))

	// Configurar concurrencia
	jobs := make(chan ParseJob, bufferSize)
	results := make(chan ParseResult, bufferSize)

	var wg sync.WaitGroup

	// Lanzar workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go ParseWorker(i, jobs, results, &wg)
	}

	// Producer: enviar jobs
	go func() {
		jobID := 0
		for i := 1; i < len(lines); i++ {
			jobs <- ParseJob{
				JobID: jobID,
				Line:  lines[i],
			}
			jobID++
		}
		close(jobs)
	}()

	// Synchronizer: cerrar canal de resultados cuando terminen todos los workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// Procesar resultados
	profiles := make([]*UserProfile, 0)
	gameNames := make(map[string]string)
	userProfiles := make(map[string]*UserProfile)

	rowsProcessed := 0
	rowsValid := 0

	for result := range results {
		rowsProcessed++

		if rowsProcessed%100000 == 0 {
			fmt.Printf("📊 Procesadas: %d filas...\n", rowsProcessed)
		}

		if !result.Valid {
			continue
		}

		rowsValid++

		// Agregar o actualizar perfil de usuario
		if existingProfile, exists := userProfiles[result.Profile.UserID]; exists {
			// Usuario ya existe, agregar juego
			existingProfile.Games[result.GameID] = result.Profile.Games[result.GameID]
		} else {
			// Nuevo usuario
			userProfiles[result.Profile.UserID] = result.Profile
			profiles = append(profiles, result.Profile)
		}

		// Agregar nombre real del juego
		gameNames[result.GameID] = result.GameName
	}

	// Finalizar métricas
	metrics = EndMetrics(metrics, int64(rowsProcessed), numWorkers)

	fmt.Printf("📊 Filas procesadas: %d\n", rowsProcessed)
	fmt.Printf("📊 Filas válidas: %d\n", rowsValid)
	fmt.Printf("📊 Usuarios únicos: %d\n", len(profiles))
	fmt.Printf("📊 Juegos únicos: %d\n", len(gameNames))

	// Mostrar métricas de rendimiento
	PrintMetrics(metrics, "MÉTRICAS DEL PARSER CONCURRENTE")

	return profiles, gameNames, nil
}

// saveProfiles guarda los perfiles
func saveProfiles(profiles []*UserProfile, filename string) {
	fmt.Printf("💾 Guardando perfiles en: %s\n", filename)

	// Crear directorio si no existe
	os.MkdirAll("data/persistence", 0755)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Error creando archivo: %v\n", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(profiles); err != nil {
		fmt.Printf("❌ Error codificando perfiles: %v\n", err)
		return
	}

	fmt.Printf("✅ Perfiles guardados exitosamente\n")
}

// saveGameNames guarda los nombres de juegos
func saveGameNames(gameNames map[string]string, filename string) {
	fmt.Printf("💾 Guardando nombres de juegos en: %s\n", filename)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Error creando archivo: %v\n", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(gameNames); err != nil {
		fmt.Printf("❌ Error codificando nombres: %v\n", err)
		return
	}

	fmt.Printf("✅ Nombres de juegos guardados exitosamente\n")
}

func runParser() {
	fmt.Printf("╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║   PARSER PARA MUESTRA DEL 10%%                             ║\n")
	fmt.Printf("║   Procesando archivo de muestra con datos reales           ║\n")
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

	inputFile := fmt.Sprintf("steam_reviews_sample_%dpct.csv", systemConfig.Sampling.Percentage)
	outputFile := "data/persistence/user_profiles_sample.gob"
	gameNamesFile := "data/persistence/game_names_sample.gob"

	// Verificar si ya existen archivos
	if fileExists(outputFile) && fileExists(gameNamesFile) {
		fmt.Printf("📁 Archivos ya existen, no es necesario procesar.\n")
		return
	}

	fmt.Printf("\n🔄 Procesando archivo: %s\n", inputFile)
	fmt.Printf("🔧 Usando %d workers para parsing\n", systemConfig.Concurrency.ParserWorkers)

	// Cargar datos con concurrencia
	profiles, gameNames, err := loadSampleCSVConcurrent(
		inputFile,
		systemConfig.Concurrency.ParserWorkers,
		systemConfig.Concurrency.BufferSize,
	)
	if err != nil {
		fmt.Printf("❌ Error cargando CSV: %v\n", err)
		return
	}

	fmt.Printf("✅ Procesamiento completado:\n")
	fmt.Printf("   - Perfiles: %d\n", len(profiles))
	fmt.Printf("   - Juegos: %d\n", len(gameNames))

	// Guardar archivos
	saveProfiles(profiles, outputFile)
	saveGameNames(gameNames, gameNamesFile)

	fmt.Printf("\n✅ Archivos guardados exitosamente!\n")
}
