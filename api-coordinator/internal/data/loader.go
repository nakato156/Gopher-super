package data

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

// LoadUserRatings lee un archivo CSV de ratings (estilo MovieLens) y
// devuelve el mapa userID -> (movieID -> rating) junto con la lista de IDs ordenada.
func LoadUserRatings(path string) (map[int]map[int]float64, []int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("abriendo archivo de ratings: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	reader.FieldsPerRecord = -1

	if _, err := reader.Read(); err != nil {
		return nil, nil, fmt.Errorf("leyendo encabezado: %w", err)
	}

	userRatings := make(map[int]map[int]float64, 1024)
	userSet := make(map[int]struct{})

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("leyendo registro de rating: %w", err)
		}
		if len(rec) < 3 {
			continue
		}

		uid, err := strconv.Atoi(rec[0])
		if err != nil {
			continue
		}
		mid, err := strconv.Atoi(rec[1])
		if err != nil {
			continue
		}
		rating, err := strconv.ParseFloat(rec[2], 64)
		if err != nil {
			continue
		}

		if _, ok := userRatings[uid]; !ok {
			userRatings[uid] = make(map[int]float64)
		}
		userRatings[uid][mid] = rating
		userSet[uid] = struct{}{}
	}

	if len(userRatings) == 0 {
		return nil, nil, fmt.Errorf("no se encontraron ratings en %s", path)
	}

	userIDs := make([]int, 0, len(userSet))
	for id := range userSet {
		userIDs = append(userIDs, id)
	}
	sort.Ints(userIDs)

	return userRatings, userIDs, nil
}
