package main

import (
	httpserver "goflix/api-coordinator/internal/http"
	"goflix/api-coordinator/internal/tcpserver"
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
