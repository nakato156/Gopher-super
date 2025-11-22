package recommend

import (
	"goflix/pkg/types"
)

// Service defines the contract for recommendation logic. For now we provide
// a simple mock implementation that returns deterministic fake IDs. Later
// this should call the recommendation engine / workers or query precomputed
// models in the database.
type Service interface {
	RecommendForUser(userID int, topN int) ([]types.Result, error)
}

type DispatchFunc func(int) ([]types.Result, error)

type recomendService struct {
	dispatch DispatchFunc
}

func NewService(dispatch DispatchFunc) Service {
	return &recomendService{dispatch: dispatch}
}

func (m *recomendService) RecommendForUser(userID int, topN int) ([]types.Result, error) {
	if m.dispatch != nil {
		results, err := m.dispatch(userID)
		if err != nil {
			return nil, err
		}
		return results, nil
	}

	return nil, nil
}
