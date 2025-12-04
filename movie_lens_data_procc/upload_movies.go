package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MovieDocument representa un documento de película en MongoDB
type MovieDocument struct {
	MovieID int      `bson:"movieId" json:"movieId"`
	Title   string   `bson:"title" json:"title"`
	Genres  []string `bson:"genres" json:"genres"`
}

// leerMoviesCSV lee el archivo de películas y devuelve un slice de MovieDocument
func leerMoviesCSV(ruta string) ([]MovieDocument, error) {
	fmt.Printf("📥 Leyendo películas desde: %s\n", ruta)

	file, err := os.Open(ruta)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Leer cabecera
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("error leyendo cabecera: %w", err)
	}

	var movies []MovieDocument
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error leyendo registro: %w", err)
		}

		if len(record) < 3 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue // Saltar IDs inválidos
		}

		title := record[1]
		genresStr := record[2]
		genres := strings.Split(genresStr, "|")

		movies = append(movies, MovieDocument{
			MovieID: id,
			Title:   title,
			Genres:  genres,
		})
	}

	fmt.Printf("✅ Leídas %d películas\n", len(movies))
	return movies, nil
}

// cargarMoviesMongoDB carga las películas a MongoDB
func cargarMoviesMongoDB(movies []MovieDocument, mongoURI, dbName, collName string) error {
	fmt.Println("\n📡 Conectando a MongoDB...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	opt := options.Client().ApplyURI(mongoURI)
	if len(mongoURI) > 10 && mongoURI[:10] == "mongodb+srv" {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opt.SetServerAPIOptions(serverAPI)
	}

	client, err := mongo.Connect(opt)
	if err != nil {
		return fmt.Errorf("error conectando a MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error haciendo ping a MongoDB: %w", err)
	}
	fmt.Println("✅ Conexión exitosa")

	collection := client.Database(dbName).Collection(collName)

	// Crear índice único por movieId
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "movieId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("⚠️ Advertencia al crear índice: %v", err)
	}

	// Verificar documentos existentes
	fmt.Println("\n🔍 Verificando películas existentes...")
	existingMovies := make(map[int]bool)
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"movieId": 1}))
	if err == nil {
		for cursor.Next(ctx) {
			var doc struct {
				MovieID int `bson:"movieId"`
			}
			if err := cursor.Decode(&doc); err == nil {
				existingMovies[doc.MovieID] = true
			}
		}
		cursor.Close(ctx)
	}

	var documents []interface{}
	skipped := 0
	for _, movie := range movies {
		if existingMovies[movie.MovieID] {
			skipped++
			continue
		}
		documents = append(documents, movie)
	}

	if skipped > 0 {
		fmt.Printf("   ⏭️  Omitidas %d películas que ya existen\n", skipped)
	}

	if len(documents) == 0 {
		fmt.Println("✅ Todas las películas ya están en la base de datos.")
		return nil
	}

	// Insertar en lotes
	batchSize := 1000
	total := len(documents)
	inserted := 0

	fmt.Printf("\n📤 Insertando %d películas nuevas en MongoDB...\n", total)

	startTime := time.Now()
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := documents[i:end]
		result, err := collection.InsertMany(ctx, batch)
		if err != nil {
			return fmt.Errorf("error insertando lote %d-%d: %w", i, end, err)
		}

		inserted += len(result.InsertedIDs)
		progress := float64(inserted) / float64(total) * 100
		fmt.Printf("\r✅ Insertadas %d/%d películas (%.1f%%)", inserted, total, progress)
	}
	fmt.Println()

	elapsed := time.Since(startTime)
	fmt.Printf("\n✅ Total de películas insertadas: %d\n", inserted)
	fmt.Printf("   Tiempo total: %.2f segundos\n", elapsed.Seconds())

	return nil
}

func main() {
	csvPath := flag.String("input", "movies.csv", "Ruta al archivo CSV de películas")
	dbName := flag.String("db", "goflixx", "Nombre de la base de datos")
	collName := flag.String("collection", "movies", "Nombre de la colección")

	defaultLocalURI := "mongodb://admin:password@localhost:27017/?authSource=admin"
	mongoURI := flag.String("uri", defaultLocalURI, "URI de conexión a MongoDB")

	flag.Parse()

	fmt.Println("🚀 Cargador de Películas a MongoDB")
	fmt.Println("=" + string(make([]byte, 40)) + "=")

	// Lógica para encontrar el archivo
	if *csvPath == "movies.csv" {
		if _, err := os.Stat(*csvPath); os.IsNotExist(err) {
			// Intentar rutas alternativas
			paths := []string{
				"../dataset/ml-latest-small/movies.csv",
				"ml-latest-small/movies.csv",
				"../dataset/ml-32m/movies.csv",
			}
			for _, p := range paths {
				if _, err := os.Stat(p); err == nil {
					*csvPath = p
					break
				}
			}
		}
	}

	if _, err := os.Stat(*csvPath); os.IsNotExist(err) {
		log.Fatalf("❌ Error: El archivo %s no existe", *csvPath)
	}

	movies, err := leerMoviesCSV(*csvPath)
	if err != nil {
		log.Fatalf("❌ Error cargando CSV: %v", err)
	}

	if err := cargarMoviesMongoDB(movies, *mongoURI, *dbName, *collName); err != nil {
		log.Fatalf("❌ Error al cargar a MongoDB: %v", err)
	}

	fmt.Printf("\n✅ Películas cargadas exitosamente!\n")
}
