package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UserRatingsDocument representa un documento de usuario en MongoDB
type UserRatingsDocument struct {
	UserID   int             `bson:"userId" json:"userId"`
	Email    string          `bson:"email" json:"email"`
	Password string          `bson:"password" json:"password"`
	Ratings  map[int]float64 `bson:"ratings" json:"ratings"`
}

// cargarMatrizDesdeJSON carga la matriz desde un archivo JSON
func cargarMatrizDesdeJSON(ruta string) (map[int]map[int]float64, error) {
	fmt.Printf("📥 Leyendo matriz desde: %s\n", ruta)

	file, err := os.Open(ruta)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo: %w", err)
	}
	defer file.Close()

	var matriz map[int]map[int]float64
	decoder := json.NewDecoder(file)

	startTime := time.Now()
	if err := decoder.Decode(&matriz); err != nil {
		return nil, fmt.Errorf("error decodificando JSON: %w", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ Matriz cargada: %d usuarios (%.2f segundos)\n", len(matriz), elapsed.Seconds())

	return matriz, nil
}

// cargarMatrizMongoDB carga la matriz a MongoDB
func cargarMatrizMongoDB(matriz map[int]map[int]float64, mongoURI, dbName, collName string) error {
	fmt.Println("\n📡 Conectando a MongoDB...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	// Conectar a MongoDB
	// ServerAPI solo es necesario para MongoDB Atlas, no para instancias locales
	opt := options.Client().ApplyURI(mongoURI)

	// Solo agregar ServerAPI si es una conexión a Atlas (mongodb+srv://)
	if len(mongoURI) > 10 && mongoURI[:10] == "mongodb+srv" {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opt.SetServerAPIOptions(serverAPI)
	}

	client, err := mongo.Connect(opt)
	if err != nil {
		return fmt.Errorf("error conectando a MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Verificar conexión
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error haciendo ping a MongoDB: %w", err)
	}
	fmt.Println("✅ Conexión exitosa")

	collection := client.Database(dbName).Collection(collName)

	// Verificar documentos existentes para evitar duplicados
	fmt.Println("\n🔍 Verificando documentos existentes...")
	existingUsers := make(map[int]bool)
	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"userId": 1}))
	if err == nil {
		for cursor.Next(ctx) {
			var doc struct {
				UserID int `bson:"userId"`
			}
			if err := cursor.Decode(&doc); err == nil {
				existingUsers[doc.UserID] = true
			}
		}
		cursor.Close(ctx)
	}

	existingCount := len(existingUsers)
	if existingCount > 0 {
		fmt.Printf("   ✅ Encontrados %d usuarios ya existentes\n", existingCount)
	}

	// Preparar documentos para inserción (solo los que no existen)
	var documents []interface{}
	skipped := 0
	for userID, ratings := range matriz {
		if existingUsers[userID] {
			skipped++
			continue
		}
		doc := UserRatingsDocument{
			UserID:   userID,
			Email:    fmt.Sprintf("user%d@example.com", userID),
			Password: fmt.Sprintf("User%d", userID),
			Ratings:  ratings,
		}
		documents = append(documents, doc)
	}

	if skipped > 0 {
		fmt.Printf("   ⏭️  Omitidos %d usuarios que ya existen\n", skipped)
	}

	if len(documents) == 0 {
		fmt.Println("✅ Todos los usuarios ya están en la base de datos. No hay nada que insertar.")
		return nil
	}

	// Insertar en lotes
	batchSize := 1000
	total := len(documents)
	inserted := 0

	fmt.Printf("\n📤 Insertando %d usuarios nuevos en MongoDB...\n", total)
	fmt.Printf("   Colección: %s.%s\n", dbName, collName)
	fmt.Printf("   Tamaño de lote: %d\n", batchSize)

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
		elapsed := time.Since(startTime)
		rate := float64(inserted) / elapsed.Seconds()
		remaining := float64(total-inserted) / rate

		fmt.Printf("✅ Insertados %d/%d usuarios (%.1f%%) - %.0f docs/s - ~%.0fs restantes\n",
			inserted, total, progress, rate, remaining)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n✅ Total de usuarios insertados: %d\n", inserted)
	fmt.Printf("   Tiempo total: %.2f segundos\n", elapsed.Seconds())
	fmt.Printf("   Velocidad promedio: %.0f documentos/segundo\n", float64(inserted)/elapsed.Seconds())

	// Verificar conteo final
	finalCount, err := collection.CountDocuments(ctx, bson.M{})
	if err == nil {
		fmt.Printf("   Total de documentos en colección: %d\n", finalCount)
	}

	return nil
}

func main() {
	// Flags de línea de comandos
	jsonPath := flag.String("input", "matriz_normalizada.json", "Ruta al archivo JSON de la matriz")
	dbName := flag.String("db", "goflixx", "Nombre de la base de datos")
	collName := flag.String("collection", "user-movie-matrix", "Nombre de la colección")

	// URI de MongoDB - por defecto usa la instancia local de Docker
	defaultLocalURI := "mongodb://admin:password@localhost:27017/?authSource=admin"
	mongoURI := flag.String("uri", defaultLocalURI, "URI de conexión a MongoDB (por defecto: instancia local)")

	flag.Parse()

	fmt.Println("🚀 Cargador de Matriz User-Item a MongoDB")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Ruta por defecto: matriz_normalizada.json en el mismo directorio donde se ejecuta el script
	// (generado por analisis.go en movie_lens_data_procc/)
	if *jsonPath == "matriz_normalizada.json" {
		// Verificar si existe en el directorio actual
		if _, err := os.Stat(*jsonPath); os.IsNotExist(err) {
			// Si no existe, intentar en el directorio padre
			*jsonPath = "../matriz_normalizada.json"
		}
	}

	// Verificar que el archivo JSON existe
	if _, err := os.Stat(*jsonPath); os.IsNotExist(err) {
		log.Fatalf("❌ Error: El archivo %s no existe", *jsonPath)
	}

	// Cargar matriz desde JSON
	matriz, err := cargarMatrizDesdeJSON(*jsonPath)
	if err != nil {
		log.Fatalf("❌ Error cargando matriz: %v", err)
	}

	if len(matriz) == 0 {
		log.Fatal("❌ La matriz está vacía")
	}

	// Mostrar información sobre la estructura de datos
	fmt.Println("\n📋 Información de la estructura de datos:")
	fmt.Printf("   Total de usuarios en la matriz: %d\n", len(matriz))

	// Mostrar ejemplo de un documento
	exampleCount := 0
	for userID, ratings := range matriz {
		exampleCount++
		fmt.Printf("\n   📄 Ejemplo de documento que se insertará:\n")
		fmt.Printf("      {\n")
		fmt.Printf("        \"userId\": %d,\n", userID)
		fmt.Printf("        \"email\": \"user%d@example.com\",\n", userID)
		fmt.Printf("        \"password\": \"User%d\",\n", userID)
		fmt.Printf("        \"ratings\": {\n")
		ratingCount := 0
		for movieID, rating := range ratings {
			if ratingCount < 3 { // Mostrar solo los primeros 3 ratings como ejemplo
				fmt.Printf("          \"%d\": %.4f,\n", movieID, rating)
			}
			ratingCount++
		}
		if ratingCount > 3 {
			fmt.Printf("          ... (y %d ratings más)\n", ratingCount-3)
		}
		fmt.Printf("        }\n")
		fmt.Printf("      }\n")
		fmt.Printf("      Total de ratings para este usuario: %d\n", ratingCount)
		if exampleCount >= 1 {
			break
		}
	}

	// Cargar a MongoDB
	fmt.Println("\n📊 Configuración de MongoDB:")
	// Ocultar credenciales en la salida
	displayURI := *mongoURI
	if len(displayURI) > 20 {
		// Ocultar contraseña en la URI para mostrar
		if idx := strings.Index(displayURI, "@"); idx > 0 {
			if userIdx := strings.Index(displayURI[:idx], "://"); userIdx > 0 {
				protocol := displayURI[:userIdx+3]
				rest := displayURI[idx:]
				displayURI = protocol + "***" + rest
			}
		}
	}
	fmt.Printf("   URI: %s\n", displayURI)
	fmt.Printf("   Base de datos: %s\n", *dbName)
	fmt.Printf("   Colección: %s\n", *collName)
	fmt.Printf("\n   📦 Estructura de cada documento:\n")
	fmt.Printf("      {\n")
	fmt.Printf("        \"userId\": <int>,           // ID del usuario\n")
	fmt.Printf("        \"email\": <string>,         // Email del usuario\n")
	fmt.Printf("        \"password\": <string>,      // Contraseña del usuario\n")
	fmt.Printf("        \"ratings\": {              // Mapa de ratings normalizados\n")
	fmt.Printf("          \"<movieId>\": <float>,   // MovieID -> Rating normalizado\n")
	fmt.Printf("          ...\n")
	fmt.Printf("        }\n")
	fmt.Printf("      }\n")

	if err := cargarMatrizMongoDB(matriz, *mongoURI, *dbName, *collName); err != nil {
		log.Fatalf("❌ Error al cargar a MongoDB: %v", err)
	}

	fmt.Printf("\n✅ Matriz cargada exitosamente en MongoDB!\n")
	fmt.Printf("   📍 Ubicación: %s.%s\n", *dbName, *collName)
	fmt.Printf("   💡 Puedes verificar con: mongosh \"%s\" --eval \"use %s; db.%s.findOne()\"\n",
		strings.Replace(*mongoURI, "password", "***", 1), *dbName, *collName)
}
