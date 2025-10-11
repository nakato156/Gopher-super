package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// FindSimilarUsersSequential encuentra usuarios similares de manera secuencial
// Esta es la versión SIN concurrencia que procesa cada comparación de usuario
// de manera secuencial, uno tras otro.
func FindSimilarUsersSequential(
	targetUser *UserProfile,
	allUsers []*UserProfile,
	config Config,
) []SimilarityResult {
	// Iniciar métricas de rendimiento
	metrics := StartMetrics()

	numJobs := len(allUsers) - 1
	similarities := make([]SimilarityResult, 0, numJobs)

	// SECUENCIAL: Procesar cada usuario uno por uno sin goroutines
	for _, user := range allUsers {
		// Saltar el mismo usuario
		if user.UserID == targetUser.UserID {
			continue
		}

		// Calcular similaridad
		score, commonGames, err := CalculateUserSimilarity(
			targetUser,
			user,
			config.Weights.Similarity,
		)

		if err != nil {
			continue
		}

		// Aplicar filtros
		passesFilter := commonGames >= config.MinCommonGames &&
			score >= config.MinSimilarityScore

		if !passesFilter {
			continue
		}

		// Agregar resultado directamente (sin canal)
		similarities = append(similarities, SimilarityResult{
			UserID:      user.UserID,
			Score:       score,
			CommonGames: commonGames,
		})
	}

	if len(similarities) == 0 {
		return []SimilarityResult{}
	}

	// Ordenar resultados por score
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

	// Finalizar métricas
	metrics = EndMetrics(metrics, int64(numJobs), 1) // 1 "worker" secuencial

	// Mostrar métricas de rendimiento
	PrintMetrics(metrics, "MÉTRICAS DE CÁLCULO DE SIMILARIDAD (SECUENCIAL)")

	return similarities
}

// RecommendGamesSequential genera recomendaciones de juegos de manera secuencial
// Esta versión utiliza FindSimilarUsersSequential en lugar de la versión concurrente
func RecommendGamesSequential(
	targetUserID string,
	allUsers []*UserProfile,
	userProfiles map[string]*UserProfile,
	gameNames map[string]string,
	config Config,
) []GameRecommendation {
	// Iniciar métricas de rendimiento
	metrics := StartMetrics()

	targetUser, exists := userProfiles[targetUserID]
	if !exists {
		return []GameRecommendation{}
	}

	// Usar la versión secuencial para encontrar usuarios similares
	similarUsers := FindSimilarUsersSequential(targetUser, allUsers, config)

	if len(similarUsers) == 0 {
		return []GameRecommendation{}
	}

	// Preparación de datos
	playedGames := make(map[string]bool)
	for gameID := range targetUser.Games {
		playedGames[gameID] = true
	}

	// Ajuste por información disponible del target
	targetGamesCount := len(targetUser.Games)
	if targetGamesCount == 0 {
		return []GameRecommendation{}
	}

	gameCounts := make(map[string]int)
	gameReasons := make(map[string]string)
	// Agregados para nueva fórmula de accuracy (predicted score)
	gameWeightSum := make(map[string]float64)      // Σ weight
	gameWeightedAdjSum := make(map[string]float64) // Σ (playtime_ajustado_normalizado * weight)
	gameContribs := make(map[string][]float64)     // contribuciones individuales (para consenso)
	// Agregados auxiliares para confidence
	gameSimSum := make(map[string]float64)         // Σ similarity
	gamePtConfSum := make(map[string]float64)      // Σ playtime_norm (para playtime_conf)
	gameRecencyConfSum := make(map[string]float64) // Σ recency

	// Cálculo de scores de recomendación
	for _, similarUser := range similarUsers {
		userProfile := userProfiles[similarUser.UserID]
		if userProfile == nil {
			continue
		}

		for gameID, playtime := range userProfile.Games {
			if !playedGames[gameID] {
				// =========================
				// ACCURACY (Predicted Score)
				// =========================
				// Defaults y extracción de features
				playtimeBase := playtime
				// recommended (binario 0/1), default 0.5
				recommended := 0.5
				if len(userProfile.Features) > 14 {
					if userProfile.Features[14] >= 0 {
						recommended = userProfile.Features[14]
					}
				}
				// playtime_at_review, default playtime_forever
				playtimeAtReview := playtimeBase
				if len(userProfile.Features) > 2 && userProfile.Features[2] > 0 {
					playtimeAtReview = userProfile.Features[2]
				}
				// playtime_last_two_weeks, default 0
				playtime2Weeks := 0.0
				if len(userProfile.Features) > 1 && userProfile.Features[1] > 0 {
					playtime2Weeks = userProfile.Features[1]
				}
				// weighted_vote_score, default 0.5
				weightedVote := 0.5
				if len(userProfile.Features) > 5 && userProfile.Features[5] > 0 {
					weightedVote = userProfile.Features[5]
				}
				// last_played timestamp, default ahora - 15552000 (≈180 días)
				nowSec := float64(time.Now().Unix())
				lastPlayed := nowSec - 15552000.0
				if len(userProfile.Features) > 8 && userProfile.Features[8] > 0 {
					lastPlayed = userProfile.Features[8]
				}
				// num_reviews, default 1
				numReviews := 1.0
				if len(userProfile.Features) > 10 && userProfile.Features[10] > 0 {
					numReviews = userProfile.Features[10]
				}

				// Bonos sobre playtime_base
				bonus := 1.0
				if playtimeBase > 6000 {
					bonus *= 1.20
				}
				if playtimeAtReview > 600 {
					bonus *= 1.15
				}
				if playtime2Weeks > 0 {
					bonus *= 1.10
				}
				if weightedVote > 0.7 {
					bonus *= 1.05
				}

				playtimeAdjusted := playtimeBase * bonus
				// Normalización a [0,1]
				ptNorm := playtimeAdjusted / 200.0
				if ptNorm > 1.0 {
					ptNorm = 1.0
				} else if ptNorm < 0.0 {
					ptNorm = 0.0
				}

				// Factores de weight
				similarity := math.Max(similarUser.Score, 0.0)
				daysSince := math.Max(0.0, (nowSec-lastPlayed)/86400.0)
				recency := math.Exp(-daysSince / 365.0)
				credibility := math.Min(1.0, numReviews/20.0)
				recFactor := 0.3
				if recommended >= 0.5 {
					recFactor = 1.0
				}
				weight := similarity * recency * credibility * recFactor

				// Agregación para predicted score
				gameWeightedAdjSum[gameID] += ptNorm * weight
				gameWeightSum[gameID] += weight
				gameCounts[gameID]++
				gameContribs[gameID] = append(gameContribs[gameID], ptNorm*weight)
				// Aux para confidence
				gameSimSum[gameID] += similarity
				gamePtConfSum[gameID] += math.Min(1.0, playtimeBase/200.0)
				gameRecencyConfSum[gameID] += recency

				gameReasons[gameID] = fmt.Sprintf("Recomendado por %s (similaridad: %.3f)",
					similarUser.UserID, similarUser.Score)
			}
		}
	}

	// Agregación y ordenamiento de recomendaciones
	recommendations := make([]GameRecommendation, 0, len(gameWeightSum))

	for gameID := range gameWeightSum {
		weightSum := gameWeightSum[gameID]
		if weightSum == 0 {
			continue
		}
		avgScore := gameWeightedAdjSum[gameID] / weightSum
		// Asegurar rango [0,1]
		avgScore = clamp01(avgScore)

		// CONFIDENCE (5 factores)
		contribs := gameContribs[gameID]
		count := float64(len(contribs))
		if count == 0 {
			continue
		}
		// 1) similarity_conf: media de similitud de contribuyentes
		similarityConf := clamp01(gameSimSum[gameID] / count)
		// 2) sample_conf: cobertura
		sampleConf := clamp01(float64(gameCounts[gameID]) / float64(len(similarUsers)))
		// 3) playtime_conf: media de playtime normalizado
		playtimeConf := clamp01(gamePtConfSum[gameID] / count)
		// 4) recency_conf: media de recency
		recencyConf := clamp01(gameRecencyConfSum[gameID] / count)
		// 5) consensus_conf: 1 - CV de contribuciones
		eps := 1e-6
		mean := avgScore
		variance := 0.0
		for _, v := range contribs {
			d := v - mean
			variance += d * d
		}
		if len(contribs) > 1 {
			variance /= float64(len(contribs) - 1)
		} else {
			variance = 0
		}
		std := math.Sqrt(variance)
		cv := std / (mean + eps)
		cvCap := 1.0 - math.Exp(-cv)
		if cvCap > 1 {
			cvCap = 1
		}
		consensusConf := 1.0 - cvCap

		wc := config.Weights.Confidence
		confidence := clamp01(
			wc.Similarity*similarityConf +
				wc.Sample*sampleConf +
				wc.Playtime*playtimeConf +
				wc.Recency*recencyConf +
				wc.Consensus*consensusConf,
		)

		if avgScore > 0.02 {
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

	if len(recommendations) == 0 {
		return []GameRecommendation{}
	}

	// Ordenamiento final
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

	// Finalizar métricas
	metrics = EndMetrics(metrics, int64(len(recommendations)), 1) // 1 worker secuencial

	// Mostrar métricas de rendimiento
	PrintMetrics(metrics, "MÉTRICAS DE GENERACIÓN DE RECOMENDACIONES (SECUENCIAL)")

	return recommendations
}

// runMotorSequential ejecuta el sistema de recomendaciones en modo secuencial
// Esta función es equivalente a runMotor() pero utiliza las versiones secuenciales
func runMotorSequential() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   SISTEMA DE RECOMENDACIONES - VERSIÓN SECUENCIAL         ║")
	fmt.Println("║   Cargando desde archivos de persistencia                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Cargar configuración
	configFile := "config.json"
	systemConfig, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Error cargando configuración: %v\n", err)
		fmt.Printf("Usando configuración por defecto\n")
		systemConfig = DefaultConfig()
	}

	// Mostrar configuración actual
	PrintConfig(systemConfig)

	// Verificar archivos de persistencia
	profilesFile := "data/persistence/user_profiles_sample.gob"
	gameNamesFile := "data/persistence/game_names_sample.gob"

	if !fileExists(profilesFile) || !fileExists(gameNamesFile) {
		fmt.Printf("❌ Archivos de persistencia no encontrados:\n")
		fmt.Printf("   - %s\n", profilesFile)
		fmt.Printf("   - %s\n", gameNamesFile)
		fmt.Printf("Ejecuta primero: go run sample_parser.go\n")
		return
	}

	// Cargar datos desde archivos de persistencia
	fmt.Printf("\nCargando datos desde archivos de persistencia...\n")

	var allUsers []*UserProfile
	var gameNames map[string]string

	// Cargar perfiles
	if err := loadGob(profilesFile, &allUsers); err != nil {
		fmt.Printf("❌ Error cargando perfiles: %v\n", err)
		return
	}

	// Cargar nombres de juegos
	if err := loadGob(gameNamesFile, &gameNames); err != nil {
		fmt.Printf("❌ Error cargando nombres de juegos: %v\n", err)
		return
	}

	fmt.Printf("Datos cargados exitosamente:\n")
	fmt.Printf("   - Perfiles: %d usuarios\n", len(allUsers))
	fmt.Printf("   - Juegos: %d juegos únicos\n", len(gameNames))

	// Crear usuario target con features completas
	fmt.Printf("\nCreando usuario target con features completas...\n")
	targetUser := createTargetFromConfig(gameNames, systemConfig.TargetGames)

	fmt.Printf("Usuario target creado con %d juegos del dataset real\n", len(targetUser.Games))

	// Agregar usuario target a la lista de usuarios reales
	allUsers = append(allUsers, targetUser)

	// Crear mapa de perfiles para búsqueda rápida
	userProfiles := make(map[string]*UserProfile)
	for _, user := range allUsers {
		userProfiles[user.UserID] = user
	}

	// Convertir configuración del sistema a configuración compatible
	config := ConvertSystemConfigToConfig(systemConfig)

	fmt.Printf("\nBuscando usuarios similares en %d usuarios (MODO SECUENCIAL)...\n", len(allUsers))

	// Ejecutar Fase 1: Encontrar usuarios similares (SECUENCIAL)
	similarityResults := FindSimilarUsersSequential(targetUser, allUsers, config)

	// Exportar matriz de similitud a CSV para revisión
	csvFile := "similarities_sequential.csv"
	if err := writeSimilaritiesCSV(targetUser.UserID, similarityResults, csvFile); err != nil {
		fmt.Printf("⚠️  Error al escribir CSV de similitud: %v\n", err)
	} else {
		fmt.Printf("CSV de similitud escrito en: %s\n", csvFile)
	}

	// Ejecutar Fase 2: Generar recomendaciones (SECUENCIAL)
	recommendations := RecommendGamesSequential(
		targetUser.UserID,
		allUsers,
		userProfiles,
		gameNames,
		config,
	)

	// Mostrar resultados finales
	fmt.Printf("\n============================================================\n")
	fmt.Printf("USUARIO TARGET: %s\n", targetUser.UserID)
	fmt.Printf("============================================================\n")

	// Mostrar usuarios similares
	if len(similarityResults) > 0 {
		fmt.Printf("\n👥 USUARIOS SIMILARES:\n")
		for i, sim := range similarityResults {
			fmt.Printf("  %d. %-20s | Similaridad: %.3f | Juegos en común: %d\n",
				i+1, sim.UserID, sim.Score, sim.CommonGames)
		}
	} else {
		fmt.Printf("\n⚠️  No se encontraron usuarios similares\n")
	}

	// Mostrar recomendaciones ordenadas por accuracy
	if len(recommendations) > 0 {
		fmt.Printf("\n🎮 JUEGOS RECOMENDADOS (ordenados por accuracy):\n")
		for i, rec := range recommendations {
			fmt.Printf("  %d. %-30s | Accuracy: %.3f | Confianza: %.3f\n",
				i+1, rec.GameName, rec.Score, rec.Confidence)
		}
	} else {
		fmt.Printf("\n⚠️  No se encontraron juegos para recomendar\n")
	}

	fmt.Printf("\n============================================================\n")
}

// CompareSequentialVsConcurrent ejecuta ambas versiones y compara resultados
func CompareSequentialVsConcurrent() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   COMPARACIÓN: SECUENCIAL vs CONCURRENTE                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Cargar configuración
	configFile := "config.json"
	systemConfig, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("❌ Error cargando configuración: %v\n", err)
		systemConfig = DefaultConfig()
	}

	// Verificar archivos de persistencia
	profilesFile := "data/persistence/user_profiles_sample.gob"
	gameNamesFile := "data/persistence/game_names_sample.gob"

	if !fileExists(profilesFile) || !fileExists(gameNamesFile) {
		fmt.Printf("❌ Archivos de persistencia no encontrados\n")
		return
	}

	// Cargar datos
	var allUsers []*UserProfile
	var gameNames map[string]string

	if err := loadGob(profilesFile, &allUsers); err != nil {
		fmt.Printf("❌ Error cargando perfiles: %v\n", err)
		return
	}

	if err := loadGob(gameNamesFile, &gameNames); err != nil {
		fmt.Printf("❌ Error cargando nombres de juegos: %v\n", err)
		return
	}

	fmt.Printf("Datos cargados: %d usuarios, %d juegos\n", len(allUsers), len(gameNames))

	// Crear usuario target
	targetUser := createTargetFromConfig(gameNames, systemConfig.TargetGames)
	allUsers = append(allUsers, targetUser)

	config := ConvertSystemConfigToConfig(systemConfig)

	fmt.Printf("\n%s\n", strings.Repeat("═", 60))
	fmt.Printf("EJECUTANDO VERSIÓN SECUENCIAL...\n")
	fmt.Printf("%s\n", strings.Repeat("═", 60))

	// Ejecutar versión secuencial
	startSeq := time.Now()
	seqResults := FindSimilarUsersSequential(targetUser, allUsers, config)
	seqDuration := time.Since(startSeq)

	fmt.Printf("\n%s\n", strings.Repeat("═", 60))
	fmt.Printf("EJECUTANDO VERSIÓN CONCURRENTE...\n")
	fmt.Printf("%s\n", strings.Repeat("═", 60))

	// Ejecutar versión concurrente
	startCon := time.Now()
	conResults := FindSimilarUsers(targetUser, allUsers, config)
	conDuration := time.Since(startCon)

	// Comparar resultados
	fmt.Printf("\n%s\n", strings.Repeat("═", 60))
	fmt.Printf(" RESULTADOS DE LA COMPARACIÓN\n")
	fmt.Printf("%s\n", strings.Repeat("═", 60))

	fmt.Printf("\n TIEMPOS DE EJECUCIÓN:\n")
	fmt.Printf("   Secuencial:   %v\n", seqDuration)
	fmt.Printf("   Concurrente:  %v (%d workers)\n", conDuration, config.NumWorkers)

	speedup := float64(seqDuration) / float64(conDuration)
	efficiency := speedup / float64(config.NumWorkers) * 100

	fmt.Printf("\nMÉTRICAS DE RENDIMIENTO:\n")
	fmt.Printf("   Speedup:      %.2fx\n", speedup)
	fmt.Printf("   Eficiencia:   %.2f%%\n", efficiency)

	fmt.Printf("\nRESULTADOS ENCONTRADOS:\n")
	fmt.Printf("   Secuencial:   %d usuarios similares\n", len(seqResults))
	fmt.Printf("   Concurrente:  %d usuarios similares\n", len(conResults))

	// Verificar que los resultados sean iguales
	fmt.Printf("\nVERIFICACIÓN DE CORRECTITUD:\n")
	if len(seqResults) == len(conResults) {
		fmt.Printf("   ✓ Mismo número de resultados\n")

		// Comparar los primeros 5 resultados
		matching := 0
		for i := 0; i < len(seqResults) && i < 5; i++ {
			if seqResults[i].UserID == conResults[i].UserID &&
				math.Abs(seqResults[i].Score-conResults[i].Score) < 0.0001 {
				matching++
			}
		}
		fmt.Printf("   ✓ %d/%d primeros resultados coinciden\n", matching,
			min(5, len(seqResults)))
	} else {
		fmt.Printf("   ⚠️  Diferente número de resultados\n")
	}

	fmt.Printf("\n%s\n", strings.Repeat("═", 60))
}

// min retorna el mínimo de dos enteros
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
