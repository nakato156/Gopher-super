package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Uso: go run . <comando>\n")
		fmt.Printf("Comandos disponibles:\n")
		fmt.Printf("  cleaner   - Limpiar y muestrear datos CSV\n")
		fmt.Printf("  parser    - Ejecutar parser de datos\n")
		fmt.Printf("  motor     - Ejecutar motor de recomendaciones\n")
		return
	}

	command := os.Args[1]

	switch command {
	case "cleaner":
		runCleaner()
	case "parser":
		runParser()
	case "motor":
		runMotor()
	default:
		fmt.Printf("Comando desconocido: %s\n", command)
		fmt.Printf("Comandos disponibles: cleaner, parser, motor\n")
	}
}
