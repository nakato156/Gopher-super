package data

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

// LoadUserRatingsFromMongo carga los ratings desde una colección de MongoDB.
// Se asume que la colección contiene documentos con la estructura:
//
//	{
//	  "userid": int,
//	  "ratigings": {
//	    "movieId": raiting_normalizado,
//	    ...
//	  }
//	}
func LoadUserRatingsFromMongo(ctx context.Context, coll *mongo.Collection) (map[int]map[int]float64, []int, error) {
	// Estructura auxiliar para decodificar el documento de MongoDB.
	// "ratigings" es un mapa donde la clave es el movieID (string) y el valor es el rating.
	type userDoc struct {
		UserID  int                `bson:"userid"`
		Ratings map[string]float64 `bson:"ratigings"`
	}

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, nil, fmt.Errorf("consultando colección mongo: %w", err)
	}
	defer cursor.Close(ctx)

	userRatings := make(map[int]map[int]float64, 1024)
	userSet := make(map[int]struct{})

	for cursor.Next(ctx) {
		var doc userDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, nil, fmt.Errorf("decodificando documento mongo: %w", err)
		}

		if _, ok := userRatings[doc.UserID]; !ok {
			userRatings[doc.UserID] = make(map[int]float64)
		}

		for movieIDStr, rating := range doc.Ratings {
			mid, err := strconv.Atoi(movieIDStr)
			if err != nil {
				// Si la clave no es un entero, la ignoramos o reportamos error.
				// En este caso, ignoramos claves mal formadas.
				continue
			}
			userRatings[doc.UserID][mid] = rating
		}
		userSet[doc.UserID] = struct{}{}
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterando cursor mongo: %w", err)
	}

	if len(userRatings) == 0 {
		return nil, nil, fmt.Errorf("no se encontraron ratings en la colección")
	}

	userIDs := make([]int, 0, len(userSet))
	for id := range userSet {
		userIDs = append(userIDs, id)
	}
	sort.Ints(userIDs)

	return userRatings, userIDs, nil
}
