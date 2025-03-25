package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"
)

type Proxy struct {
	source string
	target string
}

type SSHCommand struct {
	command string
	args    []string
	cmd     *exec.Cmd
	mu      sync.Mutex
}

func main() {
	// Configuration
	sourceAddr := "localhost:8080"
	targetAddrs := []string{"localhost:8081", "localhost:8082"}
	sshCommand := "ssh"
	sshArgs := []string{"user@remote", "some_command"}

	// Create SSH command monitor
	sshCmd := &SSHCommand{
		command: sshCommand,
		args:    sshArgs,
	}
	go sshCmd.Monitor()

	// Start TCP proxies
	var wg sync.WaitGroup
	for _, target := range targetAddrs {
		proxy := &Proxy{
			source: sourceAddr,
			target: target,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := proxy.Start(); err != nil {
				log.Printf("Proxy error: %v", err)
			}
		}()
	}
	wg.Wait()
}

func (p *Proxy) Start() error {
	listener, err := net.Listen("tcp", p.source)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("Proxy listening on %s, forwarding to %s", p.source, p.target)

	for {
		client, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %v", err)
		}

		go p.handleConnection(client)
	}
}

func (p *Proxy) handleConnection(client net.Conn) {
	defer client.Close()

	target, err := net.Dial("tcp", p.target)
	if err != nil {
		log.Printf("Failed to connect to target: %v", err)
		return
	}
	defer target.Close()

	go io.Copy(target, client)
	io.Copy(client, target)
}

func (s *SSHCommand) Monitor() {
	for {
		s.mu.Lock()
		if s.cmd == nil || s.cmd.ProcessState != nil && s.cmd.ProcessState.Exited() {
			s.startCommand()
		}
		s.mu.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func (s *SSHCommand) startCommand() {
	ctx := context.Background()
	s.cmd = exec.CommandContext(ctx, s.command, s.args...)
	
	// Capture output
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating stdout pipe: %v", err)
		return
	}
	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating stderr pipe: %v", err)
		return
	}
	go io.Copy(log.Writer(), stdout)
	go io.Copy(log.Writer(), stderr)

	if err := s.cmd.Start(); err != nil {
		log.Printf("Failed to start command: %v", err)
		return
	}
	log.Printf("Started SSH command: %s %v", s.command, s.args)
}
