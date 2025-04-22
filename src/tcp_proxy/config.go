package main

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAt         string `yaml:"listen_at" validate:"required,hostname_port"`             // Local SOCKS proxy port
	LocalSshBindTo   string `yaml:"local_ssh_bind_to" validate:"required,hostname_port"`     // Local SSH bind address
	SSHHost          string `yaml:"ssh_host" validate:"required"`                            // SSH remote host, as used in ~/.ssh/config
	SilentSshProcess bool   `yaml:"silent_ssh_process" validate:"omitempty" default:"false"` // Optional: Suppress SSH process output
	Debug            bool   `yaml:"debug" validate:"omitempty" default:"false"`              // Optional: Enable debug mode
	SSHProbePeriod   int    `yaml:"ssh_probe_period" validate:"omitempty" default:"60"`      // Optional: SSH probe period in seconds
}

func validateHostPort(hostPort string) error {
	_, _, err := net.SplitHostPort(hostPort)
	return err
}

func LoadConfig(configFile string) (*Config, error) {
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// Validate required fields
	if config.ListenAt == "" {
		return nil, fmt.Errorf("listen_at is required")
	}
	if err := validateHostPort(config.ListenAt); err != nil {
		return nil, fmt.Errorf("invalid listen_at format: %v", err)
	}

	if config.LocalSshBindTo == "" {
		return nil, fmt.Errorf("local_ssh_bind_to is required")
	}
	if err := validateHostPort(config.LocalSshBindTo); err != nil {
		return nil, fmt.Errorf("invalid local_ssh_bind_to format: %v", err)
	}

	if config.SSHHost == "" {
		return nil, fmt.Errorf("ssh_host is required")
	}

	return &config, nil
}
