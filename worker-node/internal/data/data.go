package data

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"goflix/worker-node/internal/types"
	"io"
	"os"
	"sort"
	"strconv"
)

const (
	DataDir     = "../dataset/ml-latest-small"
	RatingsFile = "ratings.csv"
	MoviesFile  = "movies.csv"
)

func mustOpen(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}

func LoadRatings(path string) ([]types.Rating, error) {
	f := mustOpen(path)
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1

	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("leyendo encabezado ratings: %w", err)
	}

	var ratings []types.Rating
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("leyendo ratings: %w", err)
		}
		if len(rec) < 3 {
			continue
		}
		uid, _ := strconv.Atoi(rec[0])
		mid, _ := strconv.Atoi(rec[1])
		val, _ := strconv.ParseFloat(rec[2], 64)
		ratings = append(ratings, types.Rating{UserID: uid, MovieID: mid, Rating: val})
	}
	return ratings, nil
}

func LoadMovieTitles(path string) (map[int]string, error) {
	f := mustOpen(path)
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.FieldsPerRecord = -1
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("leyendo encabezado movies: %w", err)
	}

	titles := make(map[int]string, 12000)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("leyendo movies: %w", err)
		}
		if len(rec) < 3 {
			continue
		}
		mid, _ := strconv.Atoi(rec[0])
		title := rec[1]
		titles[mid] = title
	}
	return titles, nil
}

func BuildIndexes(ratings []types.Rating) (map[int]map[int]float64, map[int]map[int]float64, []int, []int) {
	userRatings := make(map[int]map[int]float64, 1000)
	itemUsers := make(map[int]map[int]float64, 11000)

	userSet := make(map[int]struct{})
	itemSet := make(map[int]struct{})

	for _, rt := range ratings {
		if _, ok := userRatings[rt.UserID]; !ok {
			userRatings[rt.UserID] = make(map[int]float64)
		}
		userRatings[rt.UserID][rt.MovieID] = rt.Rating

		if _, ok := itemUsers[rt.MovieID]; !ok {
			itemUsers[rt.MovieID] = make(map[int]float64)
		}
		itemUsers[rt.MovieID][rt.UserID] = rt.Rating

		userSet[rt.UserID] = struct{}{}
		itemSet[rt.MovieID] = struct{}{}
	}

	users := make([]int, 0, len(userSet))
	for u := range userSet {
		users = append(users, u)
	}
	items := make([]int, 0, len(itemSet))
	for i := range itemSet {
		items = append(items, i)
	}
	sort.Ints(users)
	sort.Ints(items)
	return userRatings, itemUsers, users, items
}
