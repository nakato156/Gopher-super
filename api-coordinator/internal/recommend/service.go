package recommend

import (
	"context"
	"goflix/pkg/types"
	"strconv"
)

// Service defines the contract for recommendation logic.
type Service interface {
	RecommendForUser(userID int, topN int) ([]types.Result, error)
	GetPopularMovies(ctx context.Context, topN int) ([]Movie, error)
	GetRecommendationsWithDetails(ctx context.Context, userID int, topN int) ([]RecommendedMovie, error)
}

type RecommendedMovie struct {
	Movie
	Score float64 `json:"score"`
}

type DispatchFunc func(int, int) ([]types.Result, error)

type recomendService struct {
	dispatch DispatchFunc
	repo     Repository
}

func NewService(dispatch DispatchFunc, repo Repository) Service {
	return &recomendService{
		dispatch: dispatch,
		repo:     repo,
	}
}

func (m *recomendService) RecommendForUser(userID int, topN int) ([]types.Result, error) {
	if m.dispatch != nil {
		results, err := m.dispatch(userID, topN)
		if err != nil {
			return nil, err
		}
		return results, nil
	}

	return nil, nil
}

func (m *recomendService) GetPopularMovies(ctx context.Context, topN int) ([]Movie, error) {
	return m.repo.GetPopularMovies(ctx, topN)
}

func (m *recomendService) GetRecommendationsWithDetails(ctx context.Context, userID int, topN int) ([]RecommendedMovie, error) {
	// 1. Get raw recommendations
	results, err := m.RecommendForUser(userID, topN)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return []RecommendedMovie{}, nil
	}

	// Flatten results to get all neighbors
	var neighbors []types.Neighbor
	for _, res := range results {
		neighbors = append(neighbors, res.Neighbors...)
	}

	// 2. Extract IDs
	var movieIDs []int
	scoreMap := make(map[int]float64)
	for _, n := range neighbors {
		id, err := strconv.Atoi(n.ID)
		if err == nil {
			movieIDs = append(movieIDs, id)
			scoreMap[id] = n.Similarity
		}
	}

	// 3. Fetch movies
	movies, err := m.repo.GetMoviesByIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}

	// 4. Map to RecommendedMovie
	var recommended []RecommendedMovie
	for _, movie := range movies {
		score := scoreMap[movie.MovieID]
		recommended = append(recommended, RecommendedMovie{
			Movie: movie,
			Score: score,
		})
	}

	return recommended, nil
}
