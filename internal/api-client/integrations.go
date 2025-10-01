package api_client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/novuhq/novu-go/models/components"
)

const (
	integrationsPath        = "v1/integrations/"
	StrErrCreateIntegration = "error creating integration"
)

// Used because the SDK does not cleanly return 429 if the integration already exists
func (c *ApiClient) CreateIntegration(ctx context.Context, createReq *components.CreateIntegrationRequestDto) (*components.IntegrationResponseDto, *ApiClientResponse, error) {
	fetcher := fetcher[components.IntegrationResponseDto](c, http.MethodPost, integrationsPath)
	if fetcher == nil {
		tflog.Error(ctx, StrErrCreateIntegration, map[string]any{"error": ErrFetcherNil})
		return nil, nil, fmt.Errorf("%s: %w", StrErrCreateIntegration, ErrFetcherNil)
	}

	data, apiResponse, err := fetcher.Do(ctx, createReq)
	if err != nil {
		tflog.Error(ctx, StrErrCreateIntegration, map[string]any{"error": err})
		return nil, apiResponse, fmt.Errorf("%s: %w", StrErrCreateIntegration, err)
	}
	return data, apiResponse, nil
}
