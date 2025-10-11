package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
)

// ConcurrencyConfig define la configuración de concurrencia para diferentes componentes
type ConcurrencyConfig struct {
	// Configuración para el parser
	ParserWorkers int `json:"parser_workers"`

	// Configuración para el motor de recomendaciones
	SimilarityWorkers     int `json:"similarity_workers"`
	RecommendationWorkers int `json:"recommendation_workers"`

	// Configuración general
	BufferSize int `json:"buffer_size"`
}

// SamplingConfig define la configuración de muestreo de datos
type SamplingConfig struct {
	Percentage int `json:"percentage"`  // Porcentaje de datos a muestrear (1-100)
	RandomSeed int `json:"random_seed"` // Semilla para reproducibilidad
}

// SimilarityWeights define los pesos para el cálculo de similaridad entre usuarios
type SimilarityWeights struct {
	CommonGames float64 `json:"common_games"`
	Playtime    float64 `json:"playtime"`
	Reviews     float64 `json:"reviews"`
	Prefs       float64 `json:"preferences"`
}

// ConfidenceWeights define los pesos para el cálculo de confianza
type ConfidenceWeights struct {
	Similarity float64 `json:"similarity"`
	Sample     float64 `json:"sample"`
	Playtime   float64 `json:"playtime"`
	Recency    float64 `json:"recency"`
	Consensus  float64 `json:"consensus"`
}

// WeightsConfig agrupa los pesos del sistema
type WeightsConfig struct {
	Similarity SimilarityWeights `json:"similarity"`
	Confidence ConfidenceWeights `json:"confidence"`
}

// SystemConfig define la configuración completa del sistema
type SystemConfig struct {
	// Configuración de concurrencia
	Concurrency ConcurrencyConfig `json:"concurrency"`

	// Configuración de muestreo
	Sampling SamplingConfig `json:"sampling"`

	// Configuración del algoritmo de recomendaciones
	MinCommonGames     int     `json:"min_common_games"`
	MinSimilarityScore float64 `json:"min_similarity_score"`
	K                  int     `json:"k"` // Top K usuarios similares
	N                  int     `json:"n"` // Top N recomendaciones

	// Pesos del sistema
	Weights     WeightsConfig `json:"weights"`
	TargetGames []TargetGame  `json:"target_games"`
}

// DefaultConfig retorna la configuración por defecto
func DefaultConfig() SystemConfig {
	return SystemConfig{
		Concurrency: ConcurrencyConfig{
			ParserWorkers:         runtime.NumCPU(),
			SimilarityWorkers:     runtime.NumCPU() * 2,
			RecommendationWorkers: runtime.NumCPU(),
			BufferSize:            1000,
		},
		Sampling: SamplingConfig{
			Percentage: 10,
			RandomSeed: 42,
		},
		MinCommonGames:     1,
		MinSimilarityScore: 0.01,
		K:                  10,
		N:                  5,
		Weights: WeightsConfig{
			Similarity: SimilarityWeights{
				CommonGames: 0.50,
				Playtime:    0.30,
				Reviews:     0.15,
				Prefs:       0.05,
			},
			Confidence: ConfidenceWeights{
				Similarity: 0.40,
				Sample:     0.25,
				Playtime:   0.20,
				Recency:    0.10,
				Consensus:  0.05,
			},
		},
	}
}

// LoadConfig carga la configuración desde un archivo JSON
func LoadConfig(configFile string) (SystemConfig, error) {
	// Si el archivo no existe, usar configuración por defecto
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Archivo de configuración no encontrado: %s\n", configFile)
		fmt.Printf("Usando configuración por defecto\n")
		return DefaultConfig(), nil
	}

	file, err := os.Open(configFile)
	if err != nil {
		return SystemConfig{}, fmt.Errorf("error abriendo archivo de configuración: %w", err)
	}
	defer file.Close()

	var config SystemConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return SystemConfig{}, fmt.Errorf("error decodificando configuración: %w", err)
	}

	// Backfill de pesos si no están presentes en el JSON
	def := DefaultConfig()
	// Backfill para Similarity
	if config.Weights.Similarity.CommonGames == 0 && config.Weights.Similarity.Playtime == 0 &&
		config.Weights.Similarity.Reviews == 0 && config.Weights.Similarity.Prefs == 0 {
		config.Weights.Similarity = def.Weights.Similarity
	}
	// Backfill para Confidence
	if config.Weights.Confidence.Similarity == 0 && config.Weights.Confidence.Sample == 0 &&
		config.Weights.Confidence.Playtime == 0 && config.Weights.Confidence.Recency == 0 &&
		config.Weights.Confidence.Consensus == 0 {
		config.Weights.Confidence = def.Weights.Confidence
	}

	if len(config.TargetGames) > 0 {
		out := make([]TargetGame, 0, len(config.TargetGames))
		seen := make(map[string]bool)
		for _, tg := range config.TargetGames {
			if tg.ID == "" || tg.Minutes <= 0 {
				continue
			}
			if seen[tg.ID] {
				continue // evitar duplicados por ID
			}
			out = append(out, tg)
			seen[tg.ID] = true
		}
		config.TargetGames = out
	}

	// (Opcional) Fallback si target_games viene vacío
	// Si prefieres forzar que esté presente, elimina este bloque y valida fuera.
	if len(config.TargetGames) == 0 {
		fmt.Println("⚠️  target_games vacío en config; el motor creará uno ad-hoc o fallará con mensaje claro.")
		// Puedes dejarlo así, o sembrar uno mínimo aquí si lo deseas:
		// cfg.TargetGames = []TargetGame{{ID: "648800", Minutes: 120}} // Raft, ejemplo
	}

	return config, nil
}

// SaveConfig guarda la configuración en un archivo JSON
func SaveConfig(config SystemConfig, configFile string) error {
	file, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("error creando archivo de configuración: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error codificando configuración: %w", err)
	}

	return nil
}

// PrintConfig imprime la configuración actual
func PrintConfig(config SystemConfig) {
	fmt.Printf("╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    CONFIGURACIÓN ACTUAL                    ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")

	fmt.Printf("🔧 CONCURRENCIA:\n")
	fmt.Printf("   - Parser Workers: %d\n", config.Concurrency.ParserWorkers)
	fmt.Printf("   - Similarity Workers: %d\n", config.Concurrency.SimilarityWorkers)
	fmt.Printf("   - Recommendation Workers: %d\n", config.Concurrency.RecommendationWorkers)
	fmt.Printf("   - Buffer Size: %d\n", config.Concurrency.BufferSize)

	fmt.Printf("\n📊 MUESTREO:\n")
	fmt.Printf("   - Porcentaje: %d%%\n", config.Sampling.Percentage)
	fmt.Printf("   - Random Seed: %d\n", config.Sampling.RandomSeed)

	fmt.Printf("\n🎯 ALGORITMO:\n")
	fmt.Printf("   - Min Common Games: %d\n", config.MinCommonGames)
	fmt.Printf("   - Min Similarity Score: %.3f\n", config.MinSimilarityScore)
	fmt.Printf("   - Top K Users: %d\n", config.K)
	fmt.Printf("   - Top N Recommendations: %d\n", config.N)

	fmt.Printf("\n⚖️  PESOS:\n")
	fmt.Printf("   - Similarity: common=%.2f, playtime=%.2f, reviews=%.2f, prefs=%.2f\n",
		config.Weights.Similarity.CommonGames,
		config.Weights.Similarity.Playtime,
		config.Weights.Similarity.Reviews,
		config.Weights.Similarity.Prefs,
	)
	fmt.Printf("   - Confidence: similarity=%.2f, sample=%.2f, playtime=%.2f, recency=%.2f, consensus=%.2f\n",
		config.Weights.Confidence.Similarity,
		config.Weights.Confidence.Sample,
		config.Weights.Confidence.Playtime,
		config.Weights.Confidence.Recency,
		config.Weights.Confidence.Consensus,
	)

	fmt.Printf("\n💻 SISTEMA:\n")
	fmt.Printf("   - CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("   - GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}
