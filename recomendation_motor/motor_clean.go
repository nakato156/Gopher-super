package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// clamp01 limita un valor al rango [0,1]
func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// Config configuración del sistema (mantenido para compatibilidad)
type Config struct {
	MinCommonGames     int
	MinSimilarityScore float64
	K                  int
	N                  int // Top N recomendaciones
	NumWorkers         int
	BufferSize         int
	Weights            WeightsConfig
}

// ConvertSystemConfigToConfig convierte SystemConfig a Config para compatibilidad
func ConvertSystemConfigToConfig(systemConfig SystemConfig) Config {
	return Config{
		MinCommonGames:     systemConfig.MinCommonGames,
		MinSimilarityScore: systemConfig.MinSimilarityScore,
		K:                  systemConfig.K,
		N:                  systemConfig.N,
		NumWorkers:         systemConfig.Concurrency.SimilarityWorkers,
		BufferSize:         systemConfig.Concurrency.BufferSize,
		Weights:            systemConfig.Weights,
	}
}

// PearsonCorrelation calcula la correlación de Pearson
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

// CalculateUserSimilarity compara dos perfiles de usuario usando múltiples variables.
// Fórmula general (pesos runtime desde config.Weights.Similarity):
//
//	finalScore = w.common_games * commonGamesRatio
//	           + w.playtime     * playtimeSimilarity
//	           + w.reviews      * reviewSimilarity
//	           + w.preferences  * preferenceSimilarity
//
// Donde:
// - commonGamesRatio = |JuegosComunes| / |JuegosTarget|
// - playtimeSimilarity = Pearson(user1.playtime[comunes], user2.playtime[comunes]) mapeado a [0,1]
// - reviewSimilarity y preferenceSimilarity usan Features según índices documentados
// Nota: si hay < 2 juegos en común se aplica penalización proporcional (penaltyFactor) al score.
func CalculateUserSimilarity(user1, user2 *UserProfile, simWeights SimilarityWeights) (float64, int, error) {
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

	// 1. SIMILARIDAD DE JUEGOS EN COMÚN (peso desde config)
	targetGamesCount := len(user1.Games)
	commonGamesRatio := float64(numCommonGames) / float64(targetGamesCount)

	// 2. SIMILARIDAD DE COMPORTAMIENTO DE JUEGO (Peso: 0.3)
	playtimeSimilarity := calculatePlaytimeSimilarity(user1, user2, commonGames)

	// 3. SIMILARIDAD DE PATRONES DE REVIEW (Peso: 0.2)
	reviewSimilarity := calculateReviewSimilarity(user1, user2)

	// 4. SIMILARIDAD DE PREFERENCIAS (Peso: 0.1)
	preferenceSimilarity := calculatePreferenceSimilarity(user1, user2)

	// Aplicar penalización proporcional al usuario target
	penaltyFactor := 1.0
	if numCommonGames < 2 {
		penaltyFactor = 0.5 + (float64(numCommonGames)/2.0)*0.5
	}

	// Pesos desde configuración (runtime)
	// Se asume que simWeights proviene de config.Weights.Similarity
	finalScore := (commonGamesRatio * simWeights.CommonGames) +
		(playtimeSimilarity * simWeights.Playtime) +
		(reviewSimilarity * simWeights.Reviews) +
		(preferenceSimilarity * simWeights.Prefs)

	finalScore *= penaltyFactor

	// Asegurar que el score esté en el rango [0, 1]
	if finalScore > 1.0 {
		finalScore = 1.0
	} else if finalScore < 0.0 {
		finalScore = 0.0
	}

	return finalScore, numCommonGames, nil
}

// calculatePlaytimeSimilarity calcula similaridad basada en patrones de juego
func calculatePlaytimeSimilarity(user1, user2 *UserProfile, commonGames []string) float64 {
	if len(commonGames) == 0 {
		return 0.0
	}

	// Construir vectores de horas de juego para juegos en común
	vec1 := make([]float64, len(commonGames))
	vec2 := make([]float64, len(commonGames))

	for i, gameID := range commonGames {
		vec1[i] = user1.Games[gameID]
		vec2[i] = user2.Games[gameID]
	}

	// Calcular correlación de Pearson para horas de juego
	pearsonScore, err := PearsonCorrelation(vec1, vec2)
	if err != nil {
		return 0.0
	}

	// Normalizar a rango [0, 1]
	return (pearsonScore + 1.0) / 2.0
}

// calculateReviewSimilarity calcula similaridad basada en patrones de review
func calculateReviewSimilarity(user1, user2 *UserProfile) float64 {
	// Variables de review: votes_helpful, votes_funny, weighted_vote_score, comment_count
	// Índices en Features: [3, 4, 5, 11]
	reviewFeatures := []int{3, 4, 5, 11}

	similarity := 0.0
	count := 0

	for _, idx := range reviewFeatures {
		if idx < len(user1.Features) && idx < len(user2.Features) {
			// Calcular similaridad para cada feature de review
			val1, val2 := user1.Features[idx], user2.Features[idx]
			if val1 > 0 || val2 > 0 { // Solo si al menos uno tiene valor
				// Similaridad basada en la diferencia relativa
				diff := math.Abs(val1 - val2)
				maxVal := math.Max(val1, val2)
				if maxVal > 0 {
					similarity += 1.0 - (diff / maxVal)
					count++
				}
			}
		}
	}

	if count == 0 {
		return 0.5 // Valor neutral si no hay datos
	}

	return similarity / float64(count)
}

// calculatePreferenceSimilarity calcula similaridad basada en preferencias
func calculatePreferenceSimilarity(user1, user2 *UserProfile) float64 {
	// Variables de preferencias: recommended, steam_purchase, received_for_free, written_during_early_access
	// Índices en Features: [14, 15, 16, 17]
	preferenceFeatures := []int{14, 15, 16, 17}

	similarity := 0.0
	count := 0

	for _, idx := range preferenceFeatures {
		if idx < len(user1.Features) && idx < len(user2.Features) {
			val1, val2 := user1.Features[idx], user2.Features[idx]
			// Para variables binarias, 1.0 si son iguales, 0.0 si son diferentes
			if val1 == val2 {
				similarity += 1.0
			}
			count++
		}
	}

	if count == 0 {
		return 0.5 // Valor neutral si no hay datos
	}

	return similarity / float64(count)
}

// SimilarityWorker procesa jobs de similaridad
// CONCURRENCY ANALYSIS:
// - NON-CRITICAL: Cálculos de similaridad (PearsonCorrelation, CalculateUserSimilarity)
// - NON-CRITICAL: Filtros y validaciones (passesFilter)
// - CRITICAL: Escritura al canal results (results <- SimilarityResult)
func SimilarityWorker(
	workerID int,
	jobs <-chan SimilarityJob,
	results chan<- SimilarityResult,
	wg *sync.WaitGroup,
	config Config,
) {
	defer wg.Done()

	for job := range jobs {
		// NON-CRITICAL: Cálculo de similaridad (operación pura)
		score, commonGames, err := CalculateUserSimilarity(
			job.TargetUser,
			job.OtherUser,
			config.Weights.Similarity,
		)

		if err != nil {
			continue
		}

		// NON-CRITICAL: Aplicación de filtros (operación pura)
		passesFilter := commonGames >= config.MinCommonGames &&
			score >= config.MinSimilarityScore

		if !passesFilter {
			continue
		}

		// CRITICAL: Escritura al canal compartido (requiere sincronización)
		results <- SimilarityResult{
			UserID:      job.OtherUser.UserID,
			Score:       score,
			CommonGames: commonGames,
		}
	}
}

// FindSimilarUsers encuentra usuarios similares
// CONCURRENCY ANALYSIS:
// - CRITICAL: Creación y gestión de canales (jobs, results)
// - CRITICAL: Lanzamiento de goroutines (SimilarityWorker, Producer, Synchronizer)
// - CRITICAL: Sincronización con WaitGroup (wg.Add, wg.Wait)
// - CRITICAL: Cierre de canales (close(jobs), close(results))
// - CRITICAL: Agregación de resultados (append a slice compartido)
// - NON-CRITICAL: Ordenamiento final (sort.Slice)
func FindSimilarUsers(
	targetUser *UserProfile,
	allUsers []*UserProfile,
	config Config,
) []SimilarityResult {
	// Iniciar métricas de rendimiento
	metrics := StartMetrics()

	numJobs := len(allUsers) - 1

	// CRITICAL: Creación de canales compartidos
	jobs := make(chan SimilarityJob, config.BufferSize)
	results := make(chan SimilarityResult, config.BufferSize)

	var wg sync.WaitGroup

	// CRITICAL: Lanzamiento de Worker Pool (múltiples goroutines)
	for i := 0; i < config.NumWorkers; i++ {
		wg.Add(1)
		go SimilarityWorker(i, jobs, results, &wg, config)
	}

	// CRITICAL: Producer Goroutine (escritura a canal compartido)
	go func() {
		jobID := 0
		for _, user := range allUsers {
			if user.UserID == targetUser.UserID {
				continue
			}

			// CRITICAL: Escritura al canal jobs
			jobs <- SimilarityJob{
				JobID:      jobID,
				TargetUser: targetUser,
				OtherUser:  user,
			}
			jobID++
		}

		// CRITICAL: Cierre del canal jobs
		close(jobs)
	}()

	// CRITICAL: Sincronización de Workers (WaitGroup)
	go func() {
		wg.Wait()
		// CRITICAL: Cierre del canal results
		close(results)
	}()

	// CRITICAL: Agregación de Resultados (lectura de canal compartido)
	similarities := make([]SimilarityResult, 0, numJobs)

	for result := range results {
		// CRITICAL: Append a slice (operación atómica en Go, pero requiere sincronización)
		similarities = append(similarities, result)
	}

	if len(similarities) == 0 {
		return []SimilarityResult{}
	}

	// NON-CRITICAL: Post-procesamiento (operación pura)
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
	metrics = EndMetrics(metrics, int64(numJobs), config.NumWorkers)

	// Mostrar métricas de rendimiento
	PrintMetrics(metrics, "MÉTRICAS DE CÁLCULO DE SIMILARIDAD")

	return similarities
}

// RecommendGames genera recomendaciones basadas en usuarios similares
// RecommendGames genera recomendaciones de juegos
// CONCURRENCY ANALYSIS:
// - CRITICAL: Llamada a FindSimilarUsers (usa concurrencia interna)
// - NON-CRITICAL: Cálculos de scores y normalización (operaciones puras)
// - NON-CRITICAL: Agregación de datos (operaciones locales)
// - NON-CRITICAL: Ordenamiento final (sort.Slice)
func RecommendGames(
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

	// CRITICAL: Llamada a función que usa concurrencia
	similarUsers := FindSimilarUsers(targetUser, allUsers, config)

	if len(similarUsers) == 0 {
		return []GameRecommendation{}
	}

	// NON-CRITICAL: Preparación de datos (operaciones locales)
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

	// NON-CRITICAL: Cálculo de scores de recomendación (operaciones puras)
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
				// Fórmula:
				//  - playtime_ajustado = playtime_base * bonos
				//     bonos: +30% si recommended>=0.5; +20% si playtime_base>6000;
				//            +15% si playtime_at_review>600; +10% si playtime_2weeks>0;
				//            +5% si weighted_vote_score>0.7
				//  - weight = similarity * recency * credibility * rec_factor
				//     recency = exp(-dias_desde(last_played)/365)
				//     credibility = min(1, num_reviews/20)
				//     rec_factor = 1.0 si recommended>=0.5; 0.3 en caso contrario
				//  - predicted_score = Σ( norm(playtime_ajustado) * weight ) / Σ(weight)
				// Nota: norm(playtime_ajustado) se limita a [0,1] dividiendo por 200.
				// Defaults y extracción de features
				// playtime_base en minutos (del dataset); si falta, usar 0
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
				if recommended >= 0.5 {
					bonus *= 1.30
				}
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
				// Normalización a [0,1] para score usando un cap generoso (consistente con versión previa)
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

	// NON-CRITICAL: Agregación y ordenamiento de recomendaciones (operaciones locales)
	recommendations := make([]GameRecommendation, 0, len(gameWeightSum))

	for gameID := range gameWeightSum {
		// NON-CRITICAL: Cálculos de métricas (operaciones puras)
		weightSum := gameWeightSum[gameID]
		if weightSum == 0 {
			continue
		}
		avgScore := gameWeightedAdjSum[gameID] / weightSum
		// Asegurar rango [0,1]
		avgScore = clamp01(avgScore)

		// CONFIDENCE (5 factores)
		// Fórmula:
		//  confidence = 0.40*similarity_conf + 0.25*sample_conf + 0.20*playtime_conf
		//             + 0.10*recency_conf + 0.05*consensus_conf
		//  - similarity_conf: media(similarity contribuyentes) mapeada a [0,1]
		//  - sample_conf: (#contribuyentes)/(#usuarios_similares)
		//  - playtime_conf: media(playtime_forever normalizado) mapeado a [0,1]
		//  - recency_conf: media(recency de contribuyentes)
		//  - consensus_conf: 1 - coeficiente de variación de contribuciones
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
		cv := 0.0
		if mean > 0 {
			cv = std / mean
		}
		if cv > 1 {
			cv = 1
		}
		consensusConf := 1.0 - cv

		wc := config.Weights.Confidence
		confidence := clamp01(
			wc.Similarity*similarityConf +
				wc.Sample*sampleConf +
				wc.Playtime*playtimeConf +
				wc.Recency*recencyConf +
				wc.Consensus*consensusConf,
		)

		if avgScore > 0.02 { // Umbral configurable en el futuro
			gameName := gameNames[gameID]
			if gameName == "" {
				gameName = gameID
			}

			// NON-CRITICAL: Append a slice local (no compartido)
			recommendations = append(recommendations, GameRecommendation{
				GameID:     gameID,
				GameName:   gameName,
				Score:      avgScore,
				Confidence: confidence,
				Reason:     gameReasons[gameID],
			})
		}
	}

	// NON-CRITICAL: Post-procesamiento (operaciones puras)
	if len(recommendations) == 0 {
		return []GameRecommendation{}
	}

	// NON-CRITICAL: Ordenamiento final (operación pura)
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
	metrics = EndMetrics(metrics, int64(len(recommendations)), 1) // 1 worker para recomendaciones

	// Mostrar métricas de rendimiento
	PrintMetrics(metrics, "MÉTRICAS DE GENERACIÓN DE RECOMENDACIONES")

	return recommendations
}

// ============================================================================
// FUNCIONES DE PERSISTENCIA
// ============================================================================

// createTargetUserFromRealData crea un usuario target usando juegos reales del dataset
func createTargetUserFromRealData(gameNames map[string]string) *UserProfile {
	// Buscar algunos juegos populares del dataset real
	popularGames := make(map[string]float64)
	count := 0

	// Tomar los primeros 3 juegos del dataset como ejemplo
	for gameID, gameName := range gameNames {
		if count >= 3 {
			break
		}
		// Asignar horas de juego más variadas y realistas
		// Usar valores más dispersos para evitar correlaciones perfectas
		baseHours := []float64{45.0, 120.0, 85.0}
		hours := baseHours[count] + float64(count*10) // Agregar variación
		popularGames[gameID] = hours
		fmt.Printf("   - %s (ID: %s): %.0f horas\n", gameName, gameID, hours)
		count++
	}

	return &UserProfile{
		UserID:   "target_user",
		Games:    popularGames,
		Features: make([]float64, 18),
	}
}

// createTargetUserDemo crea un usuario target usando juegos concretos del dataset (por nombre)
// Intenta usar títulos populares para maximizar el solapamiento y facilitar una accuracy alta.

// createTargetUserRich crea un usuario target con Features completas y realistas
func createTargetUserRich(gameNames map[string]string) *UserProfile {
	// Seleccionar 3 juegos del dataset (primeros encontrados)
	popularGames := make(map[string]float64)
	count := 0
	for gameID, gameName := range gameNames {
		hours := []float64{140.0, 95.0, 75.0}
		if count < len(hours) {
			popularGames[gameID] = hours[count]
			fmt.Printf("   - %s (ID: %s): %.0f horas\n", gameName, gameID, hours[count])
			count++
			if count >= 3 {
				break
			}
		}
	}

	if len(popularGames) == 0 {
		return createTargetUserFromRealData(gameNames)
	}

	features := make([]float64, 18)
	now := float64(time.Now().Unix())

	// Valores realistas para probar el uso de variables complementarias
	features[0] = 120.0            // author.playtime_forever (global aproximado)
	features[1] = 180.0            // author.playtime_last_two_weeks (>0)
	features[2] = 720.0            // author.playtime_at_review (>600)
	features[3] = 12.0             // votes_helpful
	features[4] = 3.0              // votes_funny
	features[5] = 0.85             // weighted_vote_score (>0.7)
	features[6] = now - 60*24*3600 // timestamp_created (hace ~60 días)
	features[7] = now - 7*24*3600  // timestamp_updated (hace ~7 días)
	features[8] = now - 30*24*3600 // author.last_played (hace ~30 días)
	features[9] = 200.0            // author.num_games_owned
	features[10] = 25.0            // author.num_reviews (>=20 para credibilidad=1)
	features[11] = 8.0             // comment_count
	features[12] = 123456.0        // review_id
	features[13] = 0.0             // (sin uso específico)
	features[14] = 1.0             // recommended (>=0.5)
	features[15] = 1.0             // steam_purchase
	features[16] = 0.0             // received_for_free
	features[17] = 0.0             // written_during_early_access

	return &UserProfile{
		UserID:   "target_user_rich",
		Games:    popularGames,
		Features: features,
	}
}

func runMotor() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   SISTEMA DE RECOMENDACIONES CON DATOS REALES             ║")
	fmt.Println("║   Cargando desde archivos de persistencia                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

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

	// Verificar archivos de persistencia
	profilesFile := "data/persistence/user_profiles_sample.gob"
	gameNamesFile := "data/persistence/game_names_sample.gob"

	if !fileExists(profilesFile) || !fileExists(gameNamesFile) {
		fmt.Printf("❌ Archivos de persistencia no encontrados:\n")
		fmt.Printf("   - %s\n", profilesFile)
		fmt.Printf("   - %s\n", gameNamesFile)
		fmt.Printf("💡 Ejecuta primero: go run sample_parser.go\n")
		return
	}

	// Cargar datos desde archivos de persistencia
	fmt.Printf("\n📁 Cargando datos desde archivos de persistencia...\n")

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

	fmt.Printf("✅ Datos cargados exitosamente:\n")
	fmt.Printf("   - Perfiles: %d usuarios\n", len(allUsers))
	fmt.Printf("   - Juegos: %d juegos únicos\n", len(gameNames))

	// Crear usuario target con features completas para verificación
	fmt.Printf("\n🎯 Creando usuario target con features completas...\n")
	targetUser := createTargetUserRich(gameNames)

	fmt.Printf("✅ Usuario target creado con %d juegos del dataset real\n", len(targetUser.Games))

	// Agregar usuario target a la lista de usuarios reales
	allUsers = append(allUsers, targetUser)

	// Crear mapa de perfiles para búsqueda rápida
	userProfiles := make(map[string]*UserProfile)
	for _, user := range allUsers {
		userProfiles[user.UserID] = user
	}

	// Convertir configuración del sistema a configuración compatible
	config := ConvertSystemConfigToConfig(systemConfig)

	fmt.Printf("\n🔍 Buscando usuarios similares en %d usuarios...\n", len(allUsers))
	fmt.Printf("🔧 Usando %d workers para similaridad\n", config.NumWorkers)

	// Ejecutar Fase 1: Encontrar usuarios similares
	similarityResults := FindSimilarUsers(targetUser, allUsers, config)

	// Ejecutar Fase 2: Generar recomendaciones
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
