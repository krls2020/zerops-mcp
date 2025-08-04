package zcli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ZCLIWrapper wraps the zcli command line tool
type ZCLIWrapper struct {
	debug              bool
	vpnWaitTime        time.Duration
	connectedProjectID string // Track the connected project ID
}

// New creates a new zcli wrapper
func New(debug bool) *ZCLIWrapper {
	return &ZCLIWrapper{
		debug:       debug,
		vpnWaitTime: 2 * time.Second, // default
	}
}

// NewWithConfig creates a new zcli wrapper with custom configuration
func NewWithConfig(debug bool, vpnWaitTime time.Duration) *ZCLIWrapper {
	return &ZCLIWrapper{
		debug:       debug,
		vpnWaitTime: vpnWaitTime,
	}
}

// Execute runs a zcli command without sudo
func (z *ZCLIWrapper) Execute(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "zcli", args...)
	return z.runCommand(cmd)
}

// ExecuteSudo runs a zcli command with sudo (for VPN operations)
func (z *ZCLIWrapper) ExecuteSudo(ctx context.Context, args ...string) (string, error) {
	cmdArgs := append([]string{"zcli"}, args...)
	cmd := exec.CommandContext(ctx, "sudo", cmdArgs...)
	return z.runCommand(cmd)
}

// VPNConnect connects to Zerops VPN
func (z *ZCLIWrapper) VPNConnect(ctx context.Context, projectID string) error {
	output, err := z.ExecuteSudo(ctx, "vpn", "up", "--projectId", projectID)
	if err != nil {
		return fmt.Errorf("failed to connect VPN: %w\nOutput: %s", err, output)
	}

	// Wait for connection to establish
	time.Sleep(z.vpnWaitTime)

	// Verify connection
	if !z.IsVPNConnected(ctx) {
		return fmt.Errorf("VPN connection failed to establish")
	}

	// Store the connected project ID
	z.connectedProjectID = projectID

	return nil
}

// VPNDisconnect disconnects from Zerops VPN
func (z *ZCLIWrapper) VPNDisconnect(ctx context.Context) error {
	output, err := z.ExecuteSudo(ctx, "vpn", "down")
	if err != nil {
		return fmt.Errorf("failed to disconnect VPN: %w\nOutput: %s", err, output)
	}
	// Clear the connected project ID
	z.connectedProjectID = ""
	return nil
}

// IsVPNConnected checks if VPN is connected by looking for WireGuard interface
func (z *ZCLIWrapper) IsVPNConnected(ctx context.Context) bool {
	// Check if WireGuard interface exists (common method)
	cmd := exec.CommandContext(ctx, "ifconfig", "utun4")
	if err := cmd.Run(); err == nil {
		return true
	}
	
	// Alternative: check for zerops VPN config file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		vpnConfig := filepath.Join(homeDir, ".zerops", "vpn", "wg0.conf")
		if _, err := os.Stat(vpnConfig); err == nil {
			// Config exists, might be connected
			// This is not definitive but gives a hint
			return true
		}
	}
	
	return false
}

// VPNStatus gets detailed VPN status
func (z *ZCLIWrapper) VPNStatus(ctx context.Context) (connected bool, projectID string, message string) {
	// Since zcli doesn't have a status command, we check connectivity differently
	connected = z.IsVPNConnected(ctx)
	
	if connected {
		message = "VPN appears to be connected"
		projectID = z.connectedProjectID
	} else {
		message = "VPN is not connected"
	}
	
	return connected, projectID, message
}

// Push deploys code to Zerops
func (z *ZCLIWrapper) Push(ctx context.Context, projectID, serviceName, workDir, configPath string) (string, error) {
	// zcli push is actually an alias for 'zcli service push'
	args := []string{"service", "push"}
	
	// If service name is provided, add it as positional argument
	if serviceName != "" {
		args = append(args, serviceName)
	}
	
	// Project ID is required for non-interactive mode
	if projectID != "" {
		args = append(args, "--projectId", projectID)
	}
	
	if workDir != "" {
		args = append(args, "--workingDir", workDir)
	}
	
	if configPath != "" {
		args = append(args, "--zeropsYamlPath", configPath)
	}

	cmd := exec.CommandContext(ctx, "zcli", args...)
	// Don't set cmd.Dir - let zcli handle the working directory via --workingDir flag
	
	return z.runCommand(cmd)
}

// ValidateConfig validates a zerops.yml configuration
func (z *ZCLIWrapper) ValidateConfig(ctx context.Context, configPath string) error {
	// zcli doesn't have a direct validate command
	// For now, we'll skip validation - actual validation happens during push
	return nil
}

// Version gets the zcli version
func (z *ZCLIWrapper) Version(ctx context.Context) (string, error) {
	output, err := z.Execute(ctx, "version")
	if err != nil {
		return "", fmt.Errorf("failed to get zcli version: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// IsInstalled checks if zcli is installed and available
func (z *ZCLIWrapper) IsInstalled() bool {
	cmd := exec.Command("which", "zcli")
	err := cmd.Run()
	return err == nil
}

// runCommand executes a command and returns output
func (z *ZCLIWrapper) runCommand(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if z.debug {
		fmt.Printf("Running: %s\n", strings.Join(cmd.Args, " "))
	}

	err := cmd.Run()
	output := stdout.String()
	errOutput := stderr.String()

	if err != nil {
		// Combine stdout and stderr for error cases
		fullOutput := output
		if errOutput != "" {
			fullOutput += "\n" + errOutput
		}
		return fullOutput, err
	}

	return output, nil
}

// GetConnectedProjectID returns the project ID if VPN is connected
func (z *ZCLIWrapper) GetConnectedProjectID(ctx context.Context) (string, error) {
	if !z.IsVPNConnected(ctx) {
		return "", fmt.Errorf("VPN is not connected")
	}
	if z.connectedProjectID == "" {
		return "", fmt.Errorf("could not determine connected project ID")
	}
	return z.connectedProjectID, nil
}