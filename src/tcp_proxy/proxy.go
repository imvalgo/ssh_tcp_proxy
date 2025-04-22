package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Proxy struct {
	source      string
	target      string
	activeConns sync.Map
	debug       bool
}

func (p *Proxy) countActiveConns() int {
	count := 0
	p.activeConns.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func NewProxy(source, target string, debug bool) *Proxy {
	return &Proxy{
		source: source,
		target: target,
		debug:  debug,
	}
}

func (p *Proxy) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", p.source)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Close listener on context cancellation
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	log.Printf("Proxy listening on %s, forwarding to %s", p.source, p.target)

	// Start debug logging
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if p.debug {
					log.Printf("[DEBUG] Proxy %s -> %s: active", p.source, p.target)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			client, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				return fmt.Errorf("failed to accept connection: %v", err)
			}

			go p.handleConnection(client)
		}
	}
}

func (p *Proxy) handleConnection(client net.Conn) {
	connID := fmt.Sprintf("%v->%v", client.RemoteAddr(), p.target)
	p.activeConns.Store(connID, true)
	defer func() {
		client.Close()
		p.activeConns.Delete(connID)
		if p.debug {
			log.Printf("[DEBUG] Connection closed: %v", connID)
		}
	}()
	if p.debug {
		log.Printf("[DEBUG] New connection from %v", client.RemoteAddr())
	}
	target, err := net.Dial("tcp", p.target)
	if err != nil {
		log.Printf("Failed to connect to target %v: %v", p.target, err)
		return
	}

	if p.debug {
		log.Printf("[DEBUG] Connected to target %v", p.target)
	}
	defer target.Close()

	// Use WaitGroup to wait for both directions to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Target
	go func() {
		defer wg.Done()
		_, err := io.Copy(target, client)
		if err != nil {
			log.Printf("Client -> Target copy error: %v", err)
		}
		// Close write side of target connection
		if tcpConn, ok := target.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// Target -> Client
	go func() {
		defer wg.Done()
		_, err := io.Copy(client, target)
		if err != nil {
			log.Printf("Target -> Client copy error: %v", err)
		}
		// Close write side of client connection
		if tcpConn, ok := client.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// Wait for both directions to complete
	wg.Wait()
}
