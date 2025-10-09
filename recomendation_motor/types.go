package main

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

// ParseJob representa un trabajo de parsing
type ParseJob struct {
	JobID int
	Line  string
}

// ParseResult contiene el resultado de un parsing
type ParseResult struct {
	Profile  *UserProfile
	GameID   string
	GameName string
	Valid    bool
}

// PerformanceMetrics contiene métricas de rendimiento
type PerformanceMetrics struct {
	StartTime      int64   // Tiempo de inicio en nanosegundos
	EndTime        int64   // Tiempo de fin en nanosegundos
	Duration       int64   // Duración total en nanosegundos
	DurationMs     float64 // Duración en milisegundos
	DurationSec    float64 // Duración en segundos
	ItemsProcessed int64   // Número de elementos procesados
	ItemsPerSec    float64 // Elementos por segundo
	ItemsPerMs     float64 // Elementos por milisegundo
	Workers        int     // Número de workers utilizados
	Speedup        float64 // Speedup vs secuencial
	Efficiency     float64 // Eficiencia del paralelismo
	Scalability    float64 // Escalabilidad del paralelismo
}

// BenchmarkResult contiene resultados de benchmark
type BenchmarkResult struct {
	SequentialTime int64   // Tiempo secuencial
	ParallelTime   int64   // Tiempo paralelo
	Speedup        float64 // Speedup = SequentialTime / ParallelTime
	Efficiency     float64 // Eficiencia = Speedup / Workers
	Scalability    float64 // Escalabilidad (mejora por worker)
}
