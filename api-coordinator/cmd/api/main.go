package main

import (
	httpserver "goflix/api-coordinator/internal/server/http"
	tcpserver "goflix/api-coordinator/internal/server/tcp"
	"log"
	"os"
)

func main() {
	go httpserver.NewRouter()

	server := tcpserver.NewServer()
	go func() {
		for env := range server.Incoming {
			log.Printf("Mensaje recibido de %s: %s\n", env.WorkerID, env.Msg.Type)
		}
	}()
	log.Fatal(server.Start(os.Getenv("WORKER_TCP_ADDR")))
}
