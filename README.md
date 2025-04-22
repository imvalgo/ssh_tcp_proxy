# SSH SOCKS Proxy Manager

*Note: This is a small learning project created to practice Go development and test the aider AI coding assistant.*

A TCP SOCKS proxy that runs and monitors an SSH tunnel. Useful when you need:
- Persistent SOCKS5 proxy over SSH
- Monitoring and automatic reconnection
- Running on hosts with multiple network interfaces
- Containerized environments needing SSH tunneling

## Requirements

- Go 1.21+
- SSH client installed
- Proper network access to target addresses

## Configuration

Create a `config.yml` file with these required fields:
- `listen_at`: Local SOCKS proxy port (format: host:port)
- `local_ssh_bind_to`: Local address for SSH SOCKS proxy (format: host:port) 
- `ssh_host`: Remote SSH host as defined in ~/.ssh/config

Optional fields:
- `silent_ssh_process`: Suppress SSH output (default: false)
- `debug`: Enable debug logging (default: false) 
- `ssh_probe_period`: Health check interval in seconds (default: 60)

See `config.yml.example` for a complete configuration example.
