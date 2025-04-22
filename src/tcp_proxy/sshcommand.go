package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type SSHCommand struct {
	command       string
	args          []string
	cmd           *exec.Cmd
	mu            sync.Mutex
	failedChecks  int
	config        *Config
	healthChecker *HealthChecker
}

func NewSSHCommand(config *Config) *SSHCommand {
	return &SSHCommand{
		command: "ssh",
		args: []string{
			"-N",                        // No remote command
			"-D", config.LocalSshBindTo, // Dynamic SOCKS proxy on local port
			config.SSHHost,
		},
		config:        config,
		healthChecker: NewHealthChecker(config.LocalSshBindTo, config.Debug),
	}
}

func (s *SSHCommand) Monitor(ctx context.Context) {
	// Start health check ticker
	healthTicker := time.NewTicker(time.Duration(s.config.SSHProbePeriod) * time.Second)
	defer healthTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			if s.cmd != nil && s.cmd.Process != nil {
				log.Println("Shutting down SSH command...")
				syscall.Kill(-s.cmd.Process.Pid, syscall.SIGTERM)
			}
			s.mu.Unlock()
			return
		default:
		}
		s.mu.Lock()
		if s.cmd == nil || s.cmd.ProcessState != nil && s.cmd.ProcessState.Exited() {
			log.Printf("SSH command not running, attempting to start")
			s.startCommand()
		}
		// Add debug logging
		if s.config.Debug {
			if s.cmd != nil && s.cmd.ProcessState != nil {
				log.Printf("[DEBUG] SSH Command Monitor: %s %v - %s",
					s.command, s.args, s.cmd.ProcessState.String())
			} else {
				log.Printf("[DEBUG] SSH Command Monitor: %s %v - running",
					s.command, s.args)
			}
		}
		s.mu.Unlock()

		select {
		case <-healthTicker.C:
			if !s.checkHealth() {
				s.failedChecks++
				if s.failedChecks >= 3 {
					s.mu.Lock()
					log.Printf("3 failed health checks, restarting SSH")
					if s.cmd != nil && s.cmd.Process != nil {
						syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)
					}
					s.cmd = nil
					s.failedChecks = 0
					s.mu.Unlock()
				}
			} else {
				s.failedChecks = 0
			}
		default:
			// Check more frequently if the command isn't running
			if s.cmd == nil || s.cmd.ProcessState != nil && s.cmd.ProcessState.Exited() {
				time.Sleep(1 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (s *SSHCommand) checkHealth() bool {
	if s.healthChecker == nil {
		return false
	}
	return s.healthChecker.Check()
}

func (s *SSHCommand) startCommand() {
	s.cmd = exec.Command(s.command, s.args...)
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Redirect all output to /dev/null to suppress SSH output
	if s.config.SilentSshProcess {
		devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			log.Printf("Error opening /dev/null: %v", err)
			return
		}
		defer devNull.Close()

		s.cmd.Stdout = devNull
		s.cmd.Stderr = devNull
		s.cmd.Stdin = devNull
	}

	// Start the command
	if err := s.cmd.Start(); err != nil {
		log.Printf("Failed to start command: %v", err)
		return
	}
	log.Printf("Started SSH command: %s %v", s.command, s.args)

	// Wait for the command to complete in a separate goroutine
	go func() {
		err := s.cmd.Wait()
		if err != nil {
			log.Printf("Command finished with error: %v", err)
		}
		// Clean up process group
		if s.cmd.Process != nil {
			syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)
		}
	}()
}
