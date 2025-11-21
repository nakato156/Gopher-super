package recommend

// Service defines the contract for recommendation logic. For now we provide
// a simple mock implementation that returns deterministic fake IDs. Later
// this should call the recommendation engine / workers or query precomputed
// models in the database.
type Service interface {
	RecommendForUser(userID string, topN int) ([]string, error)
}

type mockService struct{}

func NewService() Service {
	return &mockService{}
}

func (m *mockService) RecommendForUser(userID string, topN int) ([]string, error) {
	res := make([]string, 0, topN)
	for i := 0; i < topN; i++ {
		res = append(res, "movie_"+userID+"_"+string(rune('0'+(i%10))))
	}
	return res, nil
}
