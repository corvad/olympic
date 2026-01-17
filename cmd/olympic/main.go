package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/corvad/olympic"
)

func main() {
	graceful := make(chan os.Signal, 1)
	signal.Notify(graceful, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-graceful
		fmt.Printf("\n")
		log.Println("Interrupt received, shutting down...")
		// Handle graceful shutdown here if needed
		olympic.Shutdown()
		os.Exit(0)
	}()
	olympic.Init("olympic.db", "testing-token-do-not-use-in-production-insecure-token")
	olympic.Run(8080)
}
