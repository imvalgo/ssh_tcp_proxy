package main

import (
	"log"
	"os"
	"sync"
)

func main() {
	log.Println("Starting TCP proxy...")
	if len(os.Args) < 2 {
		log.Fatal("Please provide path to config file as argument")
	}

	config, err := LoadConfig(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Loaded config: %+v", config)

	ctx := setupSignalHandling()
	var wg sync.WaitGroup

	log.Println("Starting proxies...")
	startProxies(ctx, config, &wg)

	log.Println("Running... (press Ctrl+C to stop)")
	wg.Wait()
	log.Println("All components stopped")
}
