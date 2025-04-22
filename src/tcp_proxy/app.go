package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func setupSignalHandling() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()
	return ctx
}

func startProxies(ctx context.Context, config *Config, wg *sync.WaitGroup) {
	// Create separate contexts for each component
	sshCtx, sshCancel := context.WithCancel(ctx)
	proxyCtx, proxyCancel := context.WithCancel(ctx)

	// Start SSH command monitor
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sshCancel()
		sshCmd := NewSSHCommand(config)
		sshCmd.Monitor(sshCtx)
	}()

	// Start TCP proxy
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer proxyCancel()
		proxy := NewProxy(config.ListenAt, config.LocalSshBindTo, config.Debug)
		if err := proxy.Start(proxyCtx); err != nil {
			log.Printf("Proxy error: %v", err)
		}
	}()

	// Wait for both components to finish
	go func() {
		wg.Wait()
		log.Println("All components stopped")
	}()
}
