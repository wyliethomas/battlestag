package network

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHResult represents the result of an SSH connectivity check
type SSHResult struct {
	Success      bool
	ResponseTime int // milliseconds
	ErrorMessage string
}

// CheckSSH tests SSH connectivity to a server
func CheckSSH(ctx context.Context, host, user string, port int, timeout time.Duration) *SSHResult {
	start := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get SSH key path
	home, err := os.UserHomeDir()
	if err != nil {
		return &SSHResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("get home directory: %v", err),
		}
	}

	keyPath := filepath.Join(home, ".ssh", "id_rsa")

	// Read private key
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return &SSHResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("read SSH key: %v", err),
		}
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return &SSHResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("parse SSH key: %v", err),
		}
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For home lab, accept any host key
		Timeout:         timeout,
	}

	// Connect
	address := fmt.Sprintf("%s:%d", host, port)

	// Use a channel to handle context cancellation
	done := make(chan struct {
		client *ssh.Client
		err    error
	}, 1)

	go func() {
		client, err := ssh.Dial("tcp", address, config)
		done <- struct {
			client *ssh.Client
			err    error
		}{client, err}
	}()

	select {
	case <-ctx.Done():
		return &SSHResult{
			Success:      false,
			ErrorMessage: "timeout",
		}
	case result := <-done:
		if result.err != nil {
			return &SSHResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("SSH connection failed: %v", result.err),
			}
		}

		// Close connection
		if result.client != nil {
			result.client.Close()
		}

		elapsed := time.Since(start)
		return &SSHResult{
			Success:      true,
			ResponseTime: int(elapsed.Milliseconds()),
		}
	}
}
