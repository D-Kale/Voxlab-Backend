package main

import (
	"log"

	"github.com/voxlab/voxlab-backend/internal/server"
)

func main() {
	srv := server.New()

	if err := srv.Init(); err != nil {
		log.Fatalf("Error inicializando servidor: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
