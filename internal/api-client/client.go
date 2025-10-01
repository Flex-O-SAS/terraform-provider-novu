package api_client

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/novuhq/novu-go/models/components"
)

var (
	defaultClientOnce sync.Once
	defaultClientInst *http.Client

	_ NovuApiClient = &ApiClient{}
)

type NovuApiClient interface {
	GetWorkflow(ctx context.Context, workflowID string) (*WorkflowResponseDto, error)
	GetWorkflowPolymorphic(ctx context.Context, workflowID string) (*WorkflowResponseDto, *ApiClientResponse, error)
	UpdateWorkflow(ctx context.Context, workflowID string, updateReq *components.UpdateWorkflowDto) (*WorkflowResponseDto, error)
	CreateWorkflow(ctx context.Context, createReq *components.CreateWorkflowDto) (*WorkflowResponseDto, error)
}

type ApiClient struct {
	client        *http.Client
	configuration ApiConfiguration
}

type ApiConfiguration struct {
	serverUrl string
	apiKey    string
}

func defaultClient() *http.Client {
	defaultClientOnce.Do(func() {
		defaultClientInst = defaultClientConfig()
	})
	return defaultClientInst
}

// use function to clone the default transport for isolation
func defaultClientConfig() *http.Client {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t = &http.Transport{}
	}

	t = t.Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = 90 * time.Second
	t.ResponseHeaderTimeout = 10 * time.Second
	t.TLSHandshakeTimeout = 10 * time.Second
	return &http.Client{
		Transport: t,
		Timeout:   time.Minute,
	}
}

func New(opts ...ApiClientOption) *ApiClient {
	apiClient := &ApiClient{
		client: defaultClient(),
		configuration: ApiConfiguration{
			serverUrl: "https://api.novu.co",
			apiKey:    "",
		},
	}
	for _, opt := range opts {
		opt(apiClient)
	}

	// mirror the novu sdk behavior to get the api key from the env if not set
	if apiClient.configuration.apiKey == "" {
		apiClient.configuration.apiKey = getApiKeyFromEnv()
	}

	return apiClient
}

func getApiKeyFromEnv() string {
	envName := getApiKeyEnvName()
	if envName == "" {
		return ""
	}
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return os.Getenv(strings.ToUpper(envName))
}

func getApiKeyEnvName() string {
	const (
		NOVU_SEC_TAG_NAME = "security"
		API_KEY_TYPE      = "apiKey"
	)
	novuSecurityType := reflect.TypeOf(components.Security{})

	for i := 0; i < novuSecurityType.NumField(); i++ {
		field := novuSecurityType.Field(i)
		tag := field.Tag.Get(NOVU_SEC_TAG_NAME)
		envName := ""
		fieldType := ""

		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		for _, part := range parts {
			keyValue := strings.Split(part, "=")
			if len(keyValue) != 2 {
				continue
			}
			switch keyValue[0] {
			case "env":
				envName = keyValue[1]
			case "type":
				fieldType = keyValue[1]
			}
		}

		// we found the apiKey type, return the env name
		if fieldType == API_KEY_TYPE {
			return envName
		}
	}
	return ""
}

// Available options for the ApiClient
type ApiClientOption func(*ApiClient)

// WithApiKey sets the API key for the ApiClient
func WithApiKey(key string) ApiClientOption {
	return func(c *ApiClient) { c.configuration.apiKey = key }
}

// WithServerUrl overrides the server URL for the ApiClient
func WithServerUrl(serverURL string) ApiClientOption {
	return func(c *ApiClient) { c.configuration.serverUrl = serverURL }
}

// WithHTTPClient overrides the HTTP client for the ApiClient
func WithHTTPClient(client *http.Client) ApiClientOption {
	return func(c *ApiClient) { c.client = client }
}
