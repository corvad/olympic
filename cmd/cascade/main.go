package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/corvad/cascade"
)

func main() {
	graceful := make(chan os.Signal, 1)
	signal.Notify(graceful, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-graceful
		fmt.Printf("\n")
		log.Println("Interrupt received, shutting down...")
		// Handle graceful shutdown here if needed
		cascade.Shutdown()
		os.Exit(0)
	}()
	dbFile := os.Getenv("DB_FILE")
	jwtSecret := os.Getenv("JWT_SECRET")
	if dbFile == "" || jwtSecret == "" {
		log.Fatal("Environment variables DB_FILE and JWT_SECRET must be set")
	}
	cascade.Init(dbFile, jwtSecret)
	cascade.Run(8080)
}
