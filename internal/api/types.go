package api

import "time"

// User represents a Zerops user
type User struct {
	ID             string         `json:"id"`
	Email          string         `json:"email"`
	FullName       string         `json:"fullName"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	Language       Language       `json:"language"`
	Status         string         `json:"status"`
	Created        time.Time      `json:"created"`
	LastUpdate     time.Time      `json:"lastUpdate"`
	ClientUserList []ClientUser   `json:"clientUserList"`
}

// Language represents user's language preference
type Language struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ClientUser represents user's association with a client
type ClientUser struct {
	ID       string        `json:"id"`
	ClientID string        `json:"clientId"`
	UserID   string        `json:"userId"`
	Status   string        `json:"status"`
	RoleCode string        `json:"roleCode"`
	Client   ClientAccount `json:"client"`
}

// ClientAccount represents a Zerops client/account
type ClientAccount struct {
	ID                      string `json:"id"`
	AccountName             string `json:"accountName"`
	PaymentProviderClientID string `json:"paymentProviderClientId"`
}

// Region represents a Zerops region
type Region struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	Address   string `json:"address"`
}

// Project represents a Zerops project
type Project struct {
	ID                  string    `json:"id"`
	ClientID            string    `json:"clientId"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Status              string    `json:"status"`
	Mode                string    `json:"mode"`
	Created             time.Time `json:"created"`
	LastUpdate          time.Time `json:"lastUpdate"`
	TagList             []string  `json:"tagList"`
	ZeropsSubdomainHost *string   `json:"zeropsSubdomainHost"`
}

// ProjectEnvironment represents project environment settings
type ProjectEnvironment struct {
	ID                  string                 `json:"id"`
	ProjectID           string                 `json:"projectId"`
	ClientID            string                 `json:"clientId"`
	ZeropsSubdomainHost string                 `json:"zeropsSubdomainHost"`
	EnvVariables        map[string]interface{} `json:"envVariables"`
}

// Service represents a Zerops service
type Service struct {
	ID                        string                 `json:"id"`
	ProjectID                 string                 `json:"projectId"`
	ClientID                  string                 `json:"clientId"`
	Name                      string                 `json:"name"`
	Status                    string                 `json:"status"`
	Mode                      string                 `json:"mode"`
	Created                   time.Time              `json:"created"`
	LastUpdate                time.Time              `json:"lastUpdate"`
	ServiceStackTypeID        string                 `json:"serviceStackTypeId"`
	ServiceStackTypeVersionID string                 `json:"serviceStackTypeVersionId"`
	ServiceStackTypeInfo      ServiceStackTypeInfo   `json:"serviceStackTypeInfo"`
	Ports                     []Port                 `json:"ports"`
	MinContainers             int                    `json:"minContainers"`
	MaxContainers             int                    `json:"maxContainers"`
	SubdomainAccess          bool                   `json:"subdomainAccess"`
	ZeropsSubdomainHost      *string                `json:"zeropsSubdomainHost"`
}

// ServiceStackTypeInfo contains service type information
type ServiceStackTypeInfo struct {
	ServiceStackTypeName        string `json:"serviceStackTypeName"`
	ServiceStackTypeCategory    string `json:"serviceStackTypeCategory"`
	ServiceStackTypeVersionName string `json:"serviceStackTypeVersionName"`
}

// Port represents a service port configuration
type Port struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Public      bool   `json:"public"`
	HTTPRouting bool   `json:"httpRouting"`
	PortRouting bool   `json:"portRouting"`
	Scheme      string `json:"scheme"`
	Description string `json:"description"`
	ServiceID   string `json:"serviceId"`
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string   `json:"name"`
	RegionID    string   `json:"regionId"`
	ClientID    string   `json:"clientId"`
	Description string   `json:"description,omitempty"`
	TagList     []string `json:"tagList"`
}

// ImportRequest represents a request to import services via YAML
type ImportRequest struct {
	ProjectID string `json:"projectId"`
	ClientID  string `json:"clientId"`
	YAML      string `json:"yaml"`
}

// SearchRequest represents a generic search request
type SearchRequest struct {
	Search []SearchFilter `json:"search,omitempty"`
	Sort   []SortCriteria `json:"sort,omitempty"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// SearchFilter represents a search filter
type SearchFilter struct {
	Name     string      `json:"name"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// SortCriteria represents sort criteria
type SortCriteria struct {
	Name      string `json:"name"`
	Ascending bool   `json:"ascending"`
}

// SearchResult represents a paginated search result
type SearchResult[T any] struct {
	Items     []T `json:"items"`
	Limit     int `json:"limit"`
	Offset    int `json:"offset"`
	TotalHits int `json:"totalHits"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Meta    []ErrorMeta `json:"meta"`
}

// ErrorMeta contains additional error metadata
type ErrorMeta struct {
	Error    string                 `json:"error"`
	Code     string                 `json:"code"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Process represents a service operation process (start/stop/restart)
type Process struct {
	ID             string       `json:"id"`
	ActionName     string       `json:"actionName"`
	Status         string       `json:"status"`
	ServiceStackID string       `json:"serviceStackId"`
	ProjectID      string       `json:"projectId"`
	ClientID       string       `json:"clientId"`
	Created        time.Time    `json:"created"`
	Started        *time.Time   `json:"started,omitempty"`
	Finished       *time.Time   `json:"finished,omitempty"`
	LastUpdate     time.Time    `json:"lastUpdate"`
	CreatedByUser  *ProcessUser `json:"createdByUser,omitempty"`
	CanceledByUser *ProcessUser `json:"canceledByUser,omitempty"`
}

// ProcessUser represents user information in a process
type ProcessUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Type     string `json:"type"`
}

// ProjectEnv represents a project-level environment variable
type ProjectEnv struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Key       string    `json:"key"`
	Content   string    `json:"content"`
	Sensitive bool      `json:"sensitive"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

// CreateProjectEnvRequest represents a request to create a project environment variable
type CreateProjectEnvRequest struct {
	ProjectID string `json:"projectId"`
	Key       string `json:"key"`
	Content   string `json:"content"`
	Sensitive bool   `json:"sensitive"`
}

// ServiceDetails extends Service with additional fields from detail endpoint
type ServiceDetails struct {
	Service
	EnvVariables    map[string]interface{} `json:"envVariables"`
	AutoScaling     *AutoScalingConfig     `json:"autoScaling"`
	VerticalScaling *VerticalScalingConfig `json:"verticalScaling"`
}

// AutoScalingConfig represents auto-scaling configuration
type AutoScalingConfig struct {
	Enabled bool `json:"enabled"`
}

// VerticalScalingConfig represents vertical scaling configuration
type VerticalScalingConfig struct {
	Enabled bool   `json:"enabled"`
	MinCPU  int    `json:"minCpu"`
	MaxCPU  int    `json:"maxCpu"`
	MinRAM  int    `json:"minRam"`
	MaxRAM  int    `json:"maxRam"`
	MinDisk int    `json:"minDisk"`
	MaxDisk int    `json:"maxDisk"`
	CPUMode string `json:"cpuMode"`
}