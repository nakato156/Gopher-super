package main

import (
	"goflix/api-coordinator/internal/tcpserver"
	"log"
	"os"
)

func main() {
	server := tcpserver.NewServer()
	go func() {
		for env := range server.Incoming {
			log.Printf("Mensaje recibido de %s: %s\n", env.WorkerID, env.Msg.Type)
		}
	}()
	log.Fatal(server.Start(os.Getenv("WORKER_TCP_ADDR")))
}
