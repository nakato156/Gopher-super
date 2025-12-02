package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

type Rating struct {
	UserID  int     `json:"userId"`
	MovieID int     `json:"movieId"`
	Rating  float64 `json:"rating"`
}

// === Función para leer el CSV ===
func leerRatingsCSV(ruta string) ([]Rating, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Saltamos la cabecera
	var ratings []Rating
	for i, r := range records {
		if i == 0 {
			continue
		}

		if len(r) < 3 || r[0] == "" || r[1] == "" || r[2] == "" {
			continue // Filtramos registros incompletos
		}

		userID, _ := strconv.Atoi(r[0])
		movieID, _ := strconv.Atoi(r[1])
		rating, _ := strconv.ParseFloat(r[2], 64)

		ratings = append(ratings, Rating{UserID: userID, MovieID: movieID, Rating: rating})
	}

	return ratings, nil
}

// === Crear la matriz usuario-película ===
func crearMatriz(ratings []Rating) map[int]map[int]float64 {
	matriz := make(map[int]map[int]float64)
	for _, r := range ratings {
		if _, ok := matriz[r.UserID]; !ok {
			matriz[r.UserID] = make(map[int]float64)
		}
		matriz[r.UserID][r.MovieID] = r.Rating
	}
	return matriz
}

// === Normalizar los ratings por usuario ===
func normalizarMatriz(matriz map[int]map[int]float64) {
	for user, pelis := range matriz {
		var suma float64
		for _, r := range pelis {
			suma += r
		}
		promedio := suma / float64(len(pelis))

		for movie, r := range pelis {
			matriz[user][movie] = r - promedio
		}
	}
}

// === Guardar matriz en JSON ===
func guardarMatrizJSON(matriz map[int]map[int]float64, ruta string) error {
	file, err := os.Create(ruta)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(matriz)
}

// === Cargar matriz desde JSON ===
func cargarMatrizJSON(ruta string) (map[int]map[int]float64, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matriz map[int]map[int]float64
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&matriz)
	return matriz, err
}

func main() {
	fmt.Println("📥 Leyendo archivo ratings.csv ...")

	// Intentar diferentes rutas posibles
	csvPath := "ml-32m/ratings.csv"
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		// Intentar en el directorio padre
		csvPath = "../ml-32m/ratings.csv"
		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			log.Fatal("Error: No se encontró ratings.csv en ml-32m/ ni en ../ml-32m/")
		}
	}

	ratings, err := leerRatingsCSV(csvPath)
	if err != nil {
		log.Fatal("Error al leer CSV:", err)
	}

	fmt.Printf("Total de registros leídos: %d\n", len(ratings))

	// === Verificar nulos ===
	if len(ratings) == 0 {
		fmt.Println("⚠️  No se encontraron datos válidos.")
		return
	}
	fmt.Println("✅ No se encontraron valores nulos o incompletos.")

	// === Crear y normalizar matriz ===
	matriz := crearMatriz(ratings)
	fmt.Println("✅ Matriz usuario–película creada correctamente.")
	fmt.Printf("Usuarios totales: %d\n", len(matriz))

	normalizarMatriz(matriz)
	fmt.Println("✅ Normalización completada.")

	// === Guardar matriz normalizada ===
	err = guardarMatrizJSON(matriz, "matriz_normalizada.json")
	if err != nil {
		log.Fatal("Error al guardar JSON:", err)
	}
	fmt.Println("💾 Matriz guardada en 'matriz_normalizada.json'")

	// === Mostrar ejemplo de usuarios ===
	contador := 0
	for user, pelis := range matriz {
		fmt.Printf("Usuario %d → %v\n", user, pelis)
		contador++
		if contador >= 2 {
			break
		}
	}

	// === Calcular promedio global de ratings normalizados ===
	var total, suma, sumaCuadrados float64
	for _, pelis := range matriz {
		for _, r := range pelis {
			suma += r
			sumaCuadrados += r * r
			total++
		}
	}
	media := suma / total
	std := math.Sqrt((sumaCuadrados / total) - (media * media))
	fmt.Printf("\nPromedio global normalizado: %.4f, Desviación estándar: %.4f\n", media, std)
}
