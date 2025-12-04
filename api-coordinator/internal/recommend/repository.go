package recommend

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Movie struct {
	MovieID int      `bson:"movieId" json:"movieId"`
	Title   string   `bson:"title" json:"title"`
	Genres  []string `bson:"genres" json:"genres"`
	Rating  float64  `bson:"rating" json:"rating"` // Average rating or score
	Views   int      `bson:"views" json:"views"`
}

type Repository interface {
	GetPopularMovies(ctx context.Context, topN int) ([]Movie, error)
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

func (r *MongoRepository) GetPopularMovies(ctx context.Context, topN int) ([]Movie, error) {
	// Aggregation pipeline to find popular movies
	// 1. Project ratings map to array of k-v pairs
	// 2. Unwind the array
	// 3. Group by movieId (k), count views, avg rating
	// 4. Sort by views (desc)
	// 5. Limit to topN
	// 6. Lookup movie details

	pipeline := mongo.Pipeline{
		{{"$project", bson.D{
			{"ratings", bson.D{{"$objectToArray", "$ratings"}}},
		}}},
		{{"$unwind", "$ratings"}},
		{{"$group", bson.D{
			{"_id", "$ratings.k"},
			{"views", bson.D{{"$sum", 1}}},
			{"avgRating", bson.D{{"$avg", "$ratings.v"}}},
		}}},
		{{"$sort", bson.D{{"views", -1}}}},
		{{"$limit", topN}},
		// Convert _id (string) to int for lookup if needed, but moviesColl uses int movieId.
		// The ratings keys are strings in the map, so we need to convert them.
		// MongoDB conversion might be tricky in aggregation if versions differ.
		// Let's assume we can lookup.
		// Actually, $lookup requires matching types. moviesColl.movieId is int.
		// ratings.k is string.
		{{"$addFields", bson.D{
			{"movieIdInt", bson.D{{"$toInt", "$_id"}}},
		}}},
		{{"$lookup", bson.D{
			{"from", r.moviesColl.Name()},
			{"localField", "movieIdInt"},
			{"foreignField", "movieId"},
			{"as", "details"},
		}}},
		{{"$unwind", "$details"}},
		{{"$project", bson.D{
			{"movieId", "$movieIdInt"},
			{"title", "$details.title"},
			{"genres", "$details.genres"},
			{"rating", "$avgRating"},
			{"views", "$views"},
		}}},
	}

	cursor, err := r.ratingsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregating popular movies: %w", err)
	}
	defer cursor.Close(ctx)

	var movies []Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, fmt.Errorf("decoding popular movies: %w", err)
	}

	return movies, nil
}
