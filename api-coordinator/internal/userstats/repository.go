package userstats

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Movie struct {
	MovieID int      `bson:"movieId" json:"movieId"`
	Title   string   `bson:"title" json:"title"`
	Genres  []string `bson:"genres" json:"genres"`
	Rating  float64  `bson:"-" json:"rating"`
}

type GenreStat struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

type Repository interface {
	GetMoviesSeen(ctx context.Context, userID int) ([]Movie, error)
	GetTopGenres(ctx context.Context, userID int) ([]GenreStat, error)
}

type MongoRepository struct {
	moviesColl  *mongo.Collection
	ratingsColl *mongo.Collection
}

func NewMongoRepository(moviesColl, ratingsColl *mongo.Collection) *MongoRepository {
	return &MongoRepository{
		moviesColl:  moviesColl,
		ratingsColl: ratingsColl,
	}
}

func (r *MongoRepository) GetMoviesSeen(ctx context.Context, userID int) ([]Movie, error) {
	// 1. Find user ratings
	var userDoc struct {
		UserID  int                `bson:"userid"`
		Ratings map[string]float64 `bson:"ratigings"`
	}

	err := r.ratingsColl.FindOne(ctx, bson.M{"userid": userID}).Decode(&userDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []Movie{}, nil
		}
		return nil, fmt.Errorf("error fetching user ratings: %w", err)
	}

	if len(userDoc.Ratings) == 0 {
		return []Movie{}, nil
	}

	// 2. Extract movie IDs and map ratings
	movieIDs := make([]int, 0, len(userDoc.Ratings))
	ratingsMap := make(map[int]float64)
	for midStr, rating := range userDoc.Ratings {
		mid, err := strconv.Atoi(midStr)
		if err == nil {
			movieIDs = append(movieIDs, mid)
			ratingsMap[mid] = rating
		}
	}

	// 3. Fetch movies
	cursor, err := r.moviesColl.Find(ctx, bson.M{"movieId": bson.M{"$in": movieIDs}})
	if err != nil {
		return nil, fmt.Errorf("error fetching movies: %w", err)
	}
	defer cursor.Close(ctx)

	var movies []Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, fmt.Errorf("error decoding movies: %w", err)
	}

	// 4. Populate ratings
	for i := range movies {
		if r, ok := ratingsMap[movies[i].MovieID]; ok {
			movies[i].Rating = r
		}
	}

	return movies, nil
}

func (r *MongoRepository) GetTopGenres(ctx context.Context, userID int) ([]GenreStat, error) {
	movies, err := r.GetMoviesSeen(ctx, userID)
	if err != nil {
		return nil, err
	}

	genreCounts := make(map[string]int)
	for _, movie := range movies {
		for _, genre := range movie.Genres {
			if genre != "(no genres listed)" {
				genreCounts[genre]++
			}
		}
	}

	var stats []GenreStat
	for genre, count := range genreCounts {
		stats = append(stats, GenreStat{Genre: genre, Count: count})
	}

	// Sort by count desc
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	// Return top 5
	if len(stats) > 5 {
		return stats[:5], nil
	}

	return stats, nil
}
