package engine

import (
	"fmt"
	"goflix/worker-node/internal/benchmark"
	"goflix/worker-node/internal/data"
	"goflix/worker-node/internal/recommend"
	"goflix/worker-node/internal/types"
	"path/filepath"
)

func Run() {
	fmt.Println("GoFlix Item-based CF (cosine) concurrente en un solo nodo")

	// TopKNeighborsN := 20
	rPath := filepath.Join(data.DataDir, data.RatingsFile)
	mPath := filepath.Join(data.DataDir, data.MoviesFile)

	fmt.Println("Cargando ratings desde:", rPath)
	ratings, err := data.LoadRatings(rPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("Cargando títulos desde:", mPath)
	titles, err := data.LoadMovieTitles(mPath)
	if err != nil {
		panic(err)
	}

	// Índices (user->items, item->users)
	fmt.Println("Construyendo índices…")
	userRatings, _, users, _ := data.BuildIndexes(ratings)
	fmt.Printf("Usuarios: %d, Ratings: %d\n", len(users), len(ratings))

	// Benchmark de paralelismo (solo similitudes)
	results, bestWorkers := benchmark.BenchmarkWorkers(userRatings, BuildSimilaritiesConcurrent)
	benchmark.PrintBench(results)
	fmt.Printf("\nMejor configuración observada: %d workers\n", bestWorkers)

	// Construir similitudes y vecindades con la mejor configuración
	fmt.Println("\nReconstruyendo similitudes con la mejor configuración…")
	sim, _ := BuildSimilaritiesConcurrent(userRatings, bestWorkers)
	fmt.Printf("Ítems con vecindad calculada: %d\n", len(sim))

	fmt.Printf("Seleccionando top-%d vecinos por ítem…\n", types.TopKNeighborsN)
	nbrs := TopKNeighbors(sim, types.TopKNeighborsN)

	// recomendar a un usuario
	if len(users) == 0 {
		fmt.Println("No hay usuarios.")
		return
	}
	demoUser := users[0]
	fmt.Printf("\nEjemplo de recomendaciones para el usuario %d\n", demoUser)
	recs := recommend.RecommendTopN(demoUser, 10, userRatings, nbrs, titles)
	if len(recs) == 0 {
		fmt.Println("No se pudieron generar recomendaciones para el usuario de ejemplo.")
		return
	}
	fmt.Println(recommend.HumanList(recs))

	// Predicción de una película candidata
	if len(recs) > 0 {
		item := recs[0].MovieID
		if p, ok := recommend.PredictForUserItem(demoUser, item, userRatings, nbrs); ok {
			fmt.Printf("Predicción para user %d sobre movie %d (%s): %.3f\n",
				demoUser, item, titles[item], p)
		}
	}
}
