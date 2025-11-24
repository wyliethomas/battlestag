package network

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// PingResult represents the result of a ping check
type PingResult struct {
	Success       bool
	ResponseTime  int // milliseconds
	ErrorMessage  string
}

// Ping performs an ICMP ping to the specified host with timeout
func Ping(ctx context.Context, host string, timeout time.Duration) *PingResult {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use ping command (works on Linux)
	// -c 1: send one packet
	// -W: timeout in seconds
	timeoutSecs := int(timeout.Seconds())
	if timeoutSecs < 1 {
		timeoutSecs = 1
	}

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", fmt.Sprintf("%d", timeoutSecs), host)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return &PingResult{
				Success:      false,
				ErrorMessage: "timeout",
			}
		}

		return &PingResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("ping failed: %v", err),
		}
	}

	// Parse response time from output
	// Example: "64 bytes from 192.168.1.1: icmp_seq=1 ttl=64 time=1.23 ms"
	re := regexp.MustCompile(`time=(\d+\.?\d*)\s*ms`)
	matches := re.FindStringSubmatch(string(output))

	responseTime := 0
	if len(matches) >= 2 {
		timeFloat, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			responseTime = int(timeFloat)
		}
	}

	return &PingResult{
		Success:      true,
		ResponseTime: responseTime,
	}
}
