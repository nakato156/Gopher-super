package recommend

import (
	"context"
	"goflix/pkg/types"
)

// Service defines the contract for recommendation logic.
type Service interface {
	RecommendForUser(userID int, topN int) ([]types.Result, error)
	GetPopularMovies(ctx context.Context, topN int) ([]Movie, error)
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
