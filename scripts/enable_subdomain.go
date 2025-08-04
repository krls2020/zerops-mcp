package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zeropsio/zerops-mcp-v3/internal/api"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ZEROPS_API_KEY")
	if apiKey == "" {
		log.Fatal("ZEROPS_API_KEY environment variable not set")
	}

	// Service ID to enable subdomain for
	serviceID := "qg67kE2LTaaUw96QPuz3cw"

	// Create API client
	client := api.NewClient(api.ClientOptions{
		BaseURL: "https://api.app-prg1.zerops.io",
		APIKey:  apiKey,
		Debug:   true,
	})

	ctx := context.Background()

	// Get service details first
	fmt.Printf("Getting service details for %s...\n", serviceID)
	service, err := client.GetService(ctx, serviceID)
	if err != nil {
		log.Fatalf("Failed to get service: %v", err)
	}

	fmt.Printf("Service: %s (Status: %s)\n", service.Name, service.Status)
	fmt.Printf("Current subdomain access: %v\n", service.SubdomainAccess)
	if service.ZeropsSubdomainHost != nil {
		fmt.Printf("Current subdomain URL: https://%s\n", *service.ZeropsSubdomainHost)
	}

	// Check if already enabled
	if service.SubdomainAccess {
		fmt.Println("✓ Subdomain access is already enabled!")
		if service.ZeropsSubdomainHost != nil {
			fmt.Printf("URL: https://%s\n", *service.ZeropsSubdomainHost)
		}
		return
	}

	// Enable subdomain access
	fmt.Println("\nEnabling subdomain access...")
	process, err := client.EnableSubdomainAccess(ctx, serviceID)
	if err != nil {
		log.Fatalf("Failed to enable subdomain access: %v", err)
	}

	fmt.Printf("Process started: %s (Status: %s)\n", process.ID, process.Status)

	// Wait for process to complete
	fmt.Println("Waiting for process to complete...")
	completedProcess, err := client.WaitForProcess(ctx, process.ID, 2*time.Minute)
	if err != nil {
		log.Fatalf("Failed to wait for process: %v", err)
	}

	fmt.Printf("Process completed with status: %s\n", completedProcess.Status)

	// Get updated service details
	fmt.Println("\nGetting updated service details...")
	updatedService, err := client.GetService(ctx, serviceID)
	if err != nil {
		log.Fatalf("Failed to get updated service: %v", err)
	}

	fmt.Printf("Updated subdomain access: %v\n", updatedService.SubdomainAccess)
	if updatedService.ZeropsSubdomainHost != nil {
		fmt.Printf("✓ Subdomain URL: https://%s\n", *updatedService.ZeropsSubdomainHost)
	} else {
		fmt.Println("⚠️  Subdomain URL not yet available")
	}
}