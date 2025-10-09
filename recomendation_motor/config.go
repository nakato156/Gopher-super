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

// AccuracyWeights define los pesos para el cálculo de accuracy
type AccuracyWeights struct {
	Time    float64 `json:"time"`
	Pref    float64 `json:"pref"`
	Recency float64 `json:"recency"`
	Alpha   float64 `json:"alpha"`
}

// ConfidenceWeights define los pesos para el cálculo de confianza
type ConfidenceWeights struct {
	Coverage  float64 `json:"coverage"`
	Agreement float64 `json:"agreement"`
	Strength  float64 `json:"strength"`
	Quality   float64 `json:"quality"`
}

// WeightsConfig agrupa los pesos del sistema
type WeightsConfig struct {
	Accuracy   AccuracyWeights   `json:"accuracy"`
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
	Weights WeightsConfig `json:"weights"`
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
			Accuracy: AccuracyWeights{
				Time:    0.4,
				Pref:    0.3,
				Recency: 0.3,
				Alpha:   1.2,
			},
			Confidence: ConfidenceWeights{
				Coverage:  0.4,
				Agreement: 0.3,
				Strength:  0.2,
				Quality:   0.1,
			},
		},
	}
}

// LoadConfig carga la configuración desde un archivo JSON
func LoadConfig(configFile string) (SystemConfig, error) {
	// Si el archivo no existe, usar configuración por defecto
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("📁 Archivo de configuración no encontrado: %s\n", configFile)
		fmt.Printf("🔧 Usando configuración por defecto\n")
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
	if config.Weights.Accuracy.Time == 0 && config.Weights.Accuracy.Pref == 0 &&
		config.Weights.Accuracy.Recency == 0 && config.Weights.Accuracy.Alpha == 0 {
		config.Weights.Accuracy = def.Weights.Accuracy
	}
	if config.Weights.Confidence.Coverage == 0 && config.Weights.Confidence.Agreement == 0 &&
		config.Weights.Confidence.Strength == 0 && config.Weights.Confidence.Quality == 0 {
		config.Weights.Confidence = def.Weights.Confidence
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
	fmt.Printf("   - Accuracy: time=%.2f, pref=%.2f, recency=%.2f, alpha=%.2f\n",
		config.Weights.Accuracy.Time,
		config.Weights.Accuracy.Pref,
		config.Weights.Accuracy.Recency,
		config.Weights.Accuracy.Alpha,
	)
	fmt.Printf("   - Confidence: coverage=%.2f, agreement=%.2f, strength=%.2f, quality=%.2f\n",
		config.Weights.Confidence.Coverage,
		config.Weights.Confidence.Agreement,
		config.Weights.Confidence.Strength,
		config.Weights.Confidence.Quality,
	)

	fmt.Printf("\n💻 SISTEMA:\n")
	fmt.Printf("   - CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("   - GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}
