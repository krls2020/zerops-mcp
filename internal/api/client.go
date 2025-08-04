package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zeropsio/zerops-mcp-v3/internal/utils"
)

// Client is the Zerops API client
type Client struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	debug       bool
	retryConfig RetryConfig
}

// ClientOptions contains options for creating a new client
type ClientOptions struct {
	BaseURL     string
	APIKey      string
	Timeout     time.Duration
	Debug       bool
	RetryConfig *RetryConfig
}

// NewClient creates a new Zerops API client
func NewClient(opts ClientOptions) *Client {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	retryConfig := DefaultRetryConfig()
	if opts.RetryConfig != nil {
		retryConfig = *opts.RetryConfig
	}

	return &Client{
		baseURL:     opts.BaseURL,
		apiKey:      opts.APIKey,
		debug:       opts.Debug,
		retryConfig: retryConfig,
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Log redirects when debugging
				if opts.Debug {
					log.Printf("[DEBUG] HTTP Redirect: %s -> %s", via[len(via)-1].URL, req.URL)
				}
				// Copy authorization header to redirected request
				if len(via) > 0 {
					req.Header.Set("Authorization", via[0].Header.Get("Authorization"))
				}
				// Allow up to 10 redirects (Go default)
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// GetBaseURL returns the base URL of the API
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// doRequest performs an HTTP request to the API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		
		if c.debug {
			log.Printf("[DEBUG] Request body: %s", utils.MaskSensitive(string(jsonBody)))
		}
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	if c.debug {
		log.Printf("[DEBUG] %s %s (API Key: %s)", method, url, utils.MaskAPIKey(c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.debug {
		log.Printf("[DEBUG] Response status: %d", resp.StatusCode)
		log.Printf("[DEBUG] Response body: %s", utils.MaskSensitive(string(respBody)))
	}
	

	// Handle error responses
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Code != "" {
			// Return structured error message
			return nil, fmt.Errorf("%d: %s - %s", resp.StatusCode, errResp.Error.Code, errResp.Error.Message)
		}
		// Fallback to generic error
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetCurrentUser gets the current user information
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/api/rest/public/user/info", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return &user, nil
}

// GetClientID returns the first available client ID for the current user
func (c *Client) GetClientID(ctx context.Context) (string, error) {
	user, err := c.GetCurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	if len(user.ClientUserList) == 0 {
		return "", fmt.Errorf("no client associations found for user")
	}

	return user.ClientUserList[0].ClientID, nil
}

// ListRegions lists all available regions
func (c *Client) ListRegions(ctx context.Context) ([]Region, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/api/rest/public/region", nil)
	if err != nil {
		return nil, err
	}

	// The response has regions in an "items" array
	var result struct {
		Items []Region `json:"items"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal regions response: %w", err)
	}

	return result.Items, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	// Ensure tagList is not nil
	if req.TagList == nil {
		req.TagList = []string{}
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/api/rest/public/project", req)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %w", err)
	}

	return &project, nil
}

// GetProject gets a specific project by ID
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", fmt.Sprintf("/api/rest/public/project/%s", projectID), nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project response: %w", err)
	}

	return &project, nil
}

// ListProjects lists projects for the current user
func (c *Client) ListProjects(ctx context.Context, clientID string) ([]Project, error) {
	searchReq := SearchRequest{
		Search: []SearchFilter{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    clientID,
			},
		},
		Sort: []SortCriteria{
			{
				Name:      "created",
				Ascending: false,
			},
		},
		Limit:  100,
		Offset: 0,
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/api/rest/public/project/search", searchReq)
	if err != nil {
		return nil, err
	}

	var result SearchResult[Project]
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects response: %w", err)
	}

	return result.Items, nil
}

// DeleteProject deletes a project and returns the process
func (c *Client) DeleteProject(ctx context.Context, projectID string) (*Process, error) {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/rest/public/project/%s", projectID), nil)
	if err != nil {
		return nil, err
	}
	
	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}
	
	return &process, nil
}

// GetProcess gets the status of a process
func (c *Client) GetProcess(ctx context.Context, processID string) (*Process, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", fmt.Sprintf("/api/rest/public/process/%s", processID), nil)
	if err != nil {
		return nil, err
	}
	
	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}
	
	return &process, nil
}

// ImportProjectServices imports services to a project via YAML
func (c *Client) ImportProjectServices(ctx context.Context, projectID, clientID, yamlData string) error {
	importReq := ImportRequest{
		ProjectID: projectID,
		ClientID:  clientID,
		YAML:      yamlData,
	}
	
	if c.debug {
		log.Printf("[DEBUG] ImportProjectServices: Sending YAML:\n%s", yamlData)
		log.Printf("[DEBUG] ImportProjectServices: Has preprocessor directive: %v", strings.Contains(yamlData, "#yamlPreprocessor=on"))
	}

	_, err := c.doRequest(ctx, "POST", "/api/rest/public/service-stack/import", importReq)
	return err
}

// ImportProject imports services to a project via YAML (wrapper for ImportProjectServices)
func (c *Client) ImportProject(ctx context.Context, req ImportRequest) error {
	// If clientID is empty, get it from the API
	clientID := req.ClientID
	if clientID == "" {
		var err error
		clientID, err = c.GetClientID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get client ID: %w", err)
		}
	}
	
	return c.ImportProjectServices(ctx, req.ProjectID, clientID, req.YAML)
}

// CreateProjectEnv creates a project-level environment variable
func (c *Client) CreateProjectEnv(ctx context.Context, projectID, key, content string, sensitive bool) (*Process, error) {
	req := CreateProjectEnvRequest{
		ProjectID: projectID,
		Key:       key,
		Content:   content,
		Sensitive: sensitive,
	}
	
	resp, err := c.doRequest(ctx, "POST", "/api/rest/public/project-env", req)
	if err != nil {
		return nil, err
	}
	
	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}
	
	return &process, nil
}

// GetProjectServices returns all services in a project
func (c *Client) GetProjectServices(ctx context.Context, projectID string) ([]Service, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", fmt.Sprintf("/api/rest/public/project/%s/service-stack", projectID), nil)
	if err != nil {
		return nil, err
	}

	var services []Service
	if err := json.Unmarshal(resp, &services); err != nil {
		return nil, fmt.Errorf("failed to unmarshal services response: %w", err)
	}

	return services, nil
}

// ListServices lists services in a project
func (c *Client) ListServices(ctx context.Context, projectID string) ([]Service, error) {
	// First, get user info to get the clientId
	user, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if len(user.ClientUserList) == 0 {
		return nil, fmt.Errorf("no client associations found for user")
	}

	// Use the first client ID
	clientID := user.ClientUserList[0].ClientID

	searchReq := SearchRequest{
		Search: []SearchFilter{
			{
				Name:     "clientId",
				Operator: "eq",
				Value:    clientID,
			},
			{
				Name:     "projectId",
				Operator: "eq",
				Value:    projectID,
			},
		},
		Sort: []SortCriteria{
			{
				Name:      "created",
				Ascending: false,
			},
		},
		Limit:  100,
		Offset: 0,
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/api/rest/public/service-stack/search", searchReq)
	if err != nil {
		return nil, err
	}

	var result SearchResult[Service]
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal services response: %w", err)
	}

	return result.Items, nil
}

// GetService gets a specific service with full details
func (c *Client) GetService(ctx context.Context, serviceID string) (*ServiceDetails, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", fmt.Sprintf("/api/rest/public/service-stack/%s", serviceID), nil)
	if err != nil {
		return nil, err
	}

	var service ServiceDetails
	if err := json.Unmarshal(resp, &service); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service response: %w", err)
	}

	return &service, nil
}

// StartService starts a service
func (c *Client) StartService(ctx context.Context, serviceID string) (*Process, error) {
	resp, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/rest/public/service-stack/%s/start", serviceID), nil)
	if err != nil {
		return nil, err
	}

	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}

	return &process, nil
}

// StopService stops a service
func (c *Client) StopService(ctx context.Context, serviceID string) (*Process, error) {
	resp, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/rest/public/service-stack/%s/stop", serviceID), nil)
	if err != nil {
		return nil, err
	}

	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}

	return &process, nil
}

// DeleteService deletes a service
func (c *Client) DeleteService(ctx context.Context, serviceID string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/rest/public/service-stack/%s", serviceID), nil)
	return err
}

// GetServiceLogs retrieves service logs
func (c *Client) GetServiceLogs(ctx context.Context, serviceID string, container string, limit int, since string) ([]string, error) {
	// First, get the service to find the project ID and check status
	if c.debug {
		log.Printf("[DEBUG] GetServiceLogs: Getting service info for ID: %s", serviceID)
	}
	service, err := c.GetService(ctx, serviceID)
	if err != nil {
		// Log the exact error from GetService
		if c.debug {
			log.Printf("[DEBUG] GetServiceLogs: GetService failed with error: %v", err)
		}
		return nil, fmt.Errorf("failed to get service info: %w", err)
	}
	
	// Check if service has been deployed
	if service.Status == "READY_TO_DEPLOY" || service.Status == "NEW" {
		return nil, fmt.Errorf("service '%s' has not been deployed yet (status: %s)", service.Name, service.Status)
	}
	
	// Step 1: Get project log access token
	path := fmt.Sprintf("/api/rest/public/project/%s/log", service.ProjectID)
	if c.debug {
		log.Printf("[DEBUG] GetServiceLogs: Getting log access token from: %s", path)
	}
	
	resp, err := c.doRequestWithRetry(ctx, "GET", path, nil)
	if err != nil {
		if c.debug {
			log.Printf("[DEBUG] GetServiceLogs: Failed to get log access token: %v", err)
		}
		return nil, fmt.Errorf("failed to get log access: %w", err)
	}
	
	// Parse the access token response
	var tokenResp struct {
		AccessToken string `json:"accessToken"`
		URL         string `json:"url"`
	}
	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse log access response: %w", err)
	}
	
	if c.debug {
		log.Printf("[DEBUG] GetServiceLogs: Got access token, proxy URL: %s", tokenResp.URL)
	}
	
	// Step 2: Fetch logs from proxy using the access token
	// Build proxy URL with query parameters
	proxyURL := fmt.Sprintf("https://proxy.app-prg1.zerops.io/api/rest/log?accessToken=%s", tokenResp.AccessToken)
	
	// Add filters for specific service
	proxyURL += fmt.Sprintf("&serviceStackId=%s", serviceID)
	
	if limit > 0 {
		proxyURL += fmt.Sprintf("&limit=%d", limit)
	}
	
	// Create new request to proxy
	proxyReq, err := http.NewRequestWithContext(ctx, "GET", proxyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}
	
	// Make request to proxy (no auth header needed, uses access token in URL)
	proxyResp, err := c.httpClient.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs from proxy: %w", err)
	}
	defer proxyResp.Body.Close()
	
	proxyBody, err := io.ReadAll(proxyResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy response: %w", err)
	}
	
	if proxyResp.StatusCode != 200 {
		return nil, fmt.Errorf("proxy returned status %d: %s", proxyResp.StatusCode, string(proxyBody))
	}
	
	// Parse the proxy response
	var proxyLogResp struct {
		Items []struct {
			ID        string    `json:"id"`
			Timestamp time.Time `json:"timestamp"`
			Hostname  string    `json:"hostname"`
			Tag       string    `json:"tag"`
			Content   string    `json:"content"`
			Message   string    `json:"message"`
			Priority  int       `json:"priority"`
			Severity  int       `json:"severity"`
		} `json:"items"`
	}
	
	if err := json.Unmarshal(proxyBody, &proxyLogResp); err != nil {
		return nil, fmt.Errorf("failed to parse proxy log response: %w", err)
	}
	
	// Convert to string array
	var logs []string
	for _, item := range proxyLogResp.Items {
		// Format: timestamp [hostname] tag: message
		logLine := fmt.Sprintf("%s [%s] %s: %s", 
			item.Timestamp.Format("2006-01-02 15:04:05"),
			item.Hostname,
			item.Tag,
			item.Message)
		logs = append(logs, logLine)
	}
	
	// If no logs found
	if len(logs) == 0 {
		return []string{}, nil
	}
	
	return logs, nil
}

// EnableSubdomainAccess enables subdomain access for a service
func (c *Client) EnableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error) {
	resp, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/rest/public/service-stack/%s/enable-subdomain-access", serviceID), nil)
	if err != nil {
		return nil, err
	}
	
	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}
	
	return &process, nil
}

// DisableSubdomainAccess disables subdomain access for a service
func (c *Client) DisableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error) {
	resp, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/rest/public/service-stack/%s/disable-subdomain-access", serviceID), nil)
	if err != nil {
		return nil, err
	}
	
	var process Process
	if err := json.Unmarshal(resp, &process); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process response: %w", err)
	}
	
	return &process, nil
}

// WaitForProcess waits for a process to complete with exponential backoff
func (c *Client) WaitForProcess(ctx context.Context, processID string, timeout time.Duration) (*Process, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Exponential backoff configuration
	minInterval := 500 * time.Millisecond
	maxInterval := 5 * time.Second
	multiplier := 1.5
	currentInterval := minInterval
	
	// Check immediately first
	process, err := c.GetProcess(ctx, processID)
	if err == nil {
		switch process.Status {
		case "SUCCESS", "FINISHED":
			return process, nil
		case "FAILED", "ERROR", "CANCELLED":
			return process, fmt.Errorf("process %s failed with status: %s", processID, process.Status)
		}
	}
	
	for {
		select {
		case <-ctx.Done():
			// Try one more time to get final status
			finalProcess, _ := c.GetProcess(context.Background(), processID)
			if finalProcess != nil {
				return finalProcess, fmt.Errorf("timeout waiting for process %s (last status: %s)", processID, finalProcess.Status)
			}
			return nil, fmt.Errorf("timeout waiting for process %s: %w", processID, ctx.Err())
			
		case <-time.After(currentInterval):
			process, err := c.GetProcess(ctx, processID)
			if err != nil {
				return nil, fmt.Errorf("failed to get process status: %w", err)
			}
			
			switch process.Status {
			case "SUCCESS", "FINISHED":
				return process, nil
			case "FAILED", "ERROR", "CANCELLED":
				return process, fmt.Errorf("process %s failed with status: %s", processID, process.Status)
			}
			
			// Increase interval with exponential backoff
			currentInterval = time.Duration(float64(currentInterval) * multiplier)
			if currentInterval > maxInterval {
				currentInterval = maxInterval
			}
		}
	}
}