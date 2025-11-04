package main

import (
	"goflix/api-coordinator/internal/server"
	"log"
)

func main() {
	server := server.NewServer()
	go func() {
		for env := range server.Incoming {
			log.Printf("Mensaje recibido de %s: %s\n", env.WorkerID, env.Msg.Type)
		}
	}()
	log.Fatal(server.Start(":9090"))
}
