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
	if jwtSecret == "" {
		log.Fatalln("Environment variables JWT_SECRET must be set")
	}
	if dbFile == "" {
		log.Fatalln("Environment variable DB_FILE must be set")
	}
	dbConfig := cascade.DBConnection{
		Address: dbFile,
	}
	kvAddress := os.Getenv("KV_ADDRESS")
	if kvAddress == "" {
		//	log.Fatalln("Environment variable KV_ADDRESS must be set")
	}
	kvConfig := cascade.KVConnection{
		Address: kvAddress,
	}
	cascade.Init(dbConfig, kvConfig, jwtSecret)
	cascade.Run(8080)
}
