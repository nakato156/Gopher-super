package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"
)

// UserProfile representa el perfil de un usuario
type UserProfile struct {
	UserID   string
	Games    map[string]float64
	Features []float64
}

// SimilarityJob representa un trabajo de comparación
type SimilarityJob struct {
	JobID      int
	TargetUser *UserProfile
	OtherUser  *UserProfile
}

// SimilarityResult contiene el resultado de una comparación
type SimilarityResult struct {
	UserID      string
	Score       float64
	CommonGames int
}

// GameRecommendation representa una recomendación de juego
type GameRecommendation struct {
	GameID     string  // ID del juego recomendado
	GameName   string  // Nombre del juego
	Score      float64 // Score de recomendación (0-1)
	Confidence float64 // Confianza en la recomendación (0-1)
	Reason     string  // Razón de la recomendación
}

// Config configuración del sistema
type Config struct {
	MinCommonGames     int
	MinSimilarityScore float64
	K                  int
	N                  int // Top N recomendaciones
	NumWorkers         int
	BufferSize         int
}

func DefaultConfig() Config {
	return Config{
		MinCommonGames:     3,
		MinSimilarityScore: 0.3,
		K:                  10,
		N:                  5, // Top 5 recomendaciones
		NumWorkers:         runtime.NumCPU() * 2,
		BufferSize:         1000,
	}
}

// PearsonCorrelation calcula la correlación de Pearson
// NON-CRITICAL SECTION: Función pura, completamente thread-safe
func PearsonCorrelation(x, y []float64) (float64, error) {
	n := len(x)

	if n != len(y) {
		return 0, fmt.Errorf("vectores deben tener la misma longitud")
	}
	if n == 0 {
		return 0, fmt.Errorf("vectores no pueden estar vacíos")
	}
	if n == 1 {
		return 0, nil
	}

	// Calcular medias
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Calcular correlación
	var numerator, sumSqDiffX, sumSqDiffY float64

	for i := 0; i < n; i++ {
		diffX := x[i] - meanX
		diffY := y[i] - meanY

		numerator += diffX * diffY
		sumSqDiffX += diffX * diffX
		sumSqDiffY += diffY * diffY
	}

	denominator := math.Sqrt(sumSqDiffX * sumSqDiffY)

	if denominator == 0 {
		return 0, nil
	}

	correlation := numerator / denominator

	if correlation > 1.0 {
		correlation = 1.0
	} else if correlation < -1.0 {
		correlation = -1.0
	}

	return correlation, nil
}

// CalculateUserSimilarity compara dos perfiles de usuario
// NON-CRITICAL SECTION: Función pura, completamente thread-safe
func CalculateUserSimilarity(user1, user2 *UserProfile) (float64, int, error) {
	// Encontrar juegos en común
	commonGames := make([]string, 0)

	for gameID := range user1.Games {
		if _, exists := user2.Games[gameID]; exists {
			commonGames = append(commonGames, gameID)
		}
	}

	numCommonGames := len(commonGames)

	if numCommonGames == 0 {
		return 0, 0, nil
	}

	// Construir vectores alineados
	vec1 := make([]float64, numCommonGames)
	vec2 := make([]float64, numCommonGames)

	for i, gameID := range commonGames {
		vec1[i] = user1.Games[gameID]
		vec2[i] = user2.Games[gameID]
	}

	// Calcular correlación de Pearson
	score, err := PearsonCorrelation(vec1, vec2)
	if err != nil {
		return 0, numCommonGames, err
	}

	return score, numCommonGames, nil
}

// SimilarityWorker procesa jobs de similaridad
// NON-CRITICAL SECTION: Esta función es completamente paralelizable
func SimilarityWorker(
	workerID int,
	jobs <-chan SimilarityJob,
	results chan<- SimilarityResult,
	wg *sync.WaitGroup,
	config Config,
) {
	defer wg.Done()

	jobsProcessed := 0
	startTime := time.Now()

	for job := range jobs {
		jobsProcessed++

		// NON-CRITICAL SECTION: Cálculo puro (thread-safe)
		score, commonGames, err := CalculateUserSimilarity(
			job.TargetUser,
			job.OtherUser,
		)

		if err != nil {
			continue
		}

		passesFilter := commonGames >= config.MinCommonGames &&
			score >= config.MinSimilarityScore

		if !passesFilter {
			continue
		}

		// CRITICAL SECTION: Comunicación vía canal (thread-safe)
		results <- SimilarityResult{
			UserID:      job.OtherUser.UserID,
			Score:       score,
			CommonGames: commonGames,
		}
	}

	// Estadísticas del worker (silencioso)
	_ = time.Since(startTime)
	_ = jobsProcessed
}

// FindSimilarUsers encuentra usuarios similares
// Esta función implementa el patrón: NON-CRITICAL -> (workers paralelos) -> CRITICAL -> NON-CRITICAL
func FindSimilarUsers(
	targetUser *UserProfile,
	allUsers []*UserProfile,
	config Config,
) []SimilarityResult {

	// NON-CRITICAL SECTION: Preparación inicial
	numJobs := len(allUsers) - 1

	jobs := make(chan SimilarityJob, config.BufferSize)
	results := make(chan SimilarityResult, config.BufferSize)

	var wg sync.WaitGroup

	// NON-CRITICAL SECTION: Lanzamiento de Worker Pool
	for i := 0; i < config.NumWorkers; i++ {
		wg.Add(1)
		go SimilarityWorker(i, jobs, results, &wg, config)
	}

	// NON-CRITICAL SECTION: Producer Goroutine
	go func() {
		jobID := 0
		for _, user := range allUsers {
			if user.UserID == targetUser.UserID {
				continue
			}

			jobs <- SimilarityJob{
				JobID:      jobID,
				TargetUser: targetUser,
				OtherUser:  user,
			}
			jobID++
		}

		close(jobs)
	}()

	// NON-CRITICAL SECTION: Sincronización de Workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// CRITICAL SECTION: Agregación de Resultados
	similarities := make([]SimilarityResult, 0, numJobs)

	for result := range results {
		similarities = append(similarities, result)
	}

	if len(similarities) == 0 {
		return []SimilarityResult{}
	}

	// NON-CRITICAL SECTION: Post-procesamiento
	sort.Slice(similarities, func(i, j int) bool {
		if similarities[i].Score == similarities[j].Score {
			return similarities[i].CommonGames > similarities[j].CommonGames
		}
		return similarities[i].Score > similarities[j].Score
	})

	// Seleccionar top K
	topK := config.K
	if len(similarities) < topK {
		topK = len(similarities)
	}
	similarities = similarities[:topK]

	return similarities
}

// RecommendGames genera recomendaciones basadas en usuarios similares
func RecommendGames(
	targetUserID string,
	allUsers []*UserProfile,
	userProfiles map[string]*UserProfile,
	gameNames map[string]string,
	config Config,
) []GameRecommendation {

	// NON-CRITICAL SECTION: Preparación inicial
	targetUser, exists := userProfiles[targetUserID]
	if !exists {
		return []GameRecommendation{}
	}

	// Ejecutar Fase 1 para encontrar usuarios similares
	similarUsers := FindSimilarUsers(targetUser, allUsers, config)

	if len(similarUsers) == 0 {
		return []GameRecommendation{}
	}

	// NON-CRITICAL SECTION: Preparar datos para predicción
	playedGames := make(map[string]bool)
	for gameID := range targetUser.Games {
		playedGames[gameID] = true
	}

	gameScores := make(map[string]float64)
	gameCounts := make(map[string]int)
	gameReasons := make(map[string]string)

	// NON-CRITICAL SECTION: Cálculo de scores de recomendación
	for _, similarUser := range similarUsers {
		userProfile := userProfiles[similarUser.UserID]
		if userProfile == nil {
			continue
		}

		for gameID, playtime := range userProfile.Games {
			if !playedGames[gameID] {
				normalizedPlaytime := playtime / 100.0
				recommendationScore := similarUser.Score * normalizedPlaytime

				gameScores[gameID] += recommendationScore
				gameCounts[gameID]++

				gameReasons[gameID] = fmt.Sprintf("Recomendado por %s (similaridad: %.3f)",
					similarUser.UserID, similarUser.Score)
			}
		}
	}

	// CRITICAL SECTION: Agregación y ordenamiento de recomendaciones
	recommendations := make([]GameRecommendation, 0, len(gameScores))

	for gameID, totalScore := range gameScores {
		avgScore := totalScore / float64(gameCounts[gameID])
		confidence := math.Min(float64(gameCounts[gameID])/float64(len(similarUsers)), 1.0)

		if avgScore > 0.1 {
			gameName := gameNames[gameID]
			if gameName == "" {
				gameName = gameID
			}

			recommendations = append(recommendations, GameRecommendation{
				GameID:     gameID,
				GameName:   gameName,
				Score:      avgScore,
				Confidence: confidence,
				Reason:     gameReasons[gameID],
			})
		}
	}

	// NON-CRITICAL SECTION: Post-procesamiento
	if len(recommendations) == 0 {
		return []GameRecommendation{}
	}

	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Score == recommendations[j].Score {
			return recommendations[i].Confidence > recommendations[j].Confidence
		}
		return recommendations[i].Score > recommendations[j].Score
	})

	topN := config.N
	if len(recommendations) < topN {
		topN = len(recommendations)
	}
	recommendations = recommendations[:topN]

	return recommendations
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   SISTEMA DE RECOMENDACIONES COMPLETO                     ║")
	fmt.Println("║   Fase 1: Similaridad + Fase 2: Predicción              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	config := DefaultConfig()

	// Usuario objetivo
	targetUser := &UserProfile{
		UserID: "user_target",
		Games: map[string]float64{
			"portal_2":       100.0,
			"terraria":       80.0,
			"stardew_valley": 50.0,
			"hollow_knight":  120.0,
		},
	}

	// Otros usuarios
	allUsers := []*UserProfile{
		targetUser,
		{
			UserID: "user_001_similar",
			Games: map[string]float64{
				"portal_2":       110.0,
				"terraria":       85.0,
				"stardew_valley": 55.0,
				"hollow_knight":  130.0,
				"celeste":        200.0,
			},
		},
		{
			UserID: "user_002_very_similar",
			Games: map[string]float64{
				"portal_2":       105.0,
				"terraria":       82.0,
				"stardew_valley": 52.0,
				"hollow_knight":  125.0,
				"hades":          180.0,
			},
		},
		{
			UserID: "user_003_opposite",
			Games: map[string]float64{
				"portal_2":       20.0,
				"terraria":       15.0,
				"stardew_valley": 10.0,
				"call_of_duty":   500.0,
			},
		},
		{
			UserID: "user_004_few_common",
			Games: map[string]float64{
				"portal_2":  95.0,
				"minecraft": 300.0,
				"fortnite":  400.0,
			},
		},
		{
			UserID: "user_005_different",
			Games: map[string]float64{
				"fifa":   200.0,
				"nba2k":  150.0,
				"madden": 100.0,
			},
		},
	}

	similarUsers := FindSimilarUsers(targetUser, allUsers, config)

	// Crear mapas para Fase 2
	userProfiles := make(map[string]*UserProfile)
	for _, user := range allUsers {
		userProfiles[user.UserID] = user
	}

	// Nombres de juegos
	gameNames := map[string]string{
		"portal_2":       "Portal 2",
		"terraria":       "Terraria",
		"stardew_valley": "Stardew Valley",
		"hollow_knight":  "Hollow Knight",
		"celeste":        "Celeste",
		"hades":          "Hades",
		"call_of_duty":   "Call of Duty",
		"minecraft":      "Minecraft",
		"fortnite":       "Fortnite",
		"fifa":           "FIFA",
		"nba2k":          "NBA 2K",
		"madden":         "Madden NFL",
	}

	// Ejecutar Fase 2: Predicción de recomendaciones
	recommendations := RecommendGames(
		targetUser.UserID,
		allUsers,
		userProfiles,
		gameNames,
		config,
	)

	// Mostrar resultados finales
	fmt.Printf("\n============================================================\n")
	fmt.Printf("🎯 USUARIO TARGET: %s\n", targetUser.UserID)
	fmt.Printf("============================================================\n")

	// Mostrar usuarios similares
	fmt.Printf("\n👥 USUARIOS SIMILARES:\n")
	for i, sim := range similarUsers {
		fmt.Printf("  %d. %-20s | Similaridad: %.3f | Juegos en común: %d\n",
			i+1, sim.UserID, sim.Score, sim.CommonGames)
	}

	// Mostrar recomendaciones ordenadas por accuracy
	if len(recommendations) > 0 {
		fmt.Printf("\n🎮 JUEGOS RECOMENDADOS (ordenados por accuracy):\n")
		for i, rec := range recommendations {
			fmt.Printf("  %d. %-20s | Accuracy: %.3f | Confianza: %.3f\n",
				i+1, rec.GameName, rec.Score, rec.Confidence)
		}
	} else {
		fmt.Printf("\n⚠️  No se encontraron juegos para recomendar\n")
	}

	fmt.Printf("\n============================================================\n")
}
