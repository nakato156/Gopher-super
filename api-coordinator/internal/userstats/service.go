package userstats

import "context"

type Service interface {
	GetMoviesSeen(ctx context.Context, userID int) ([]Movie, error)
	GetTopGenres(ctx context.Context, userID int) ([]GenreStat, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetMoviesSeen(ctx context.Context, userID int) ([]Movie, error) {
	return s.repo.GetMoviesSeen(ctx, userID)
}

func (s *service) GetTopGenres(ctx context.Context, userID int) ([]GenreStat, error) {
	return s.repo.GetTopGenres(ctx, userID)
}
