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
	StrErrUpdateIntegration = "error updating integration"
)

// SDK Issue : if the integration already exists, the SDK tries again in a loop instead of returning a 429 error
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

func (c *ApiClient) UpdateIntegration(ctx context.Context, integrationID string, updateReq *components.UpdateIntegrationRequestDto) (*components.IntegrationResponseDto, *ApiClientResponse, error) {
	fetcher := fetcher[components.IntegrationResponseDto](c, http.MethodPut, integrationsPath+integrationID)
	if fetcher == nil {
		tflog.Error(ctx, StrErrUpdateIntegration, map[string]any{"error": ErrFetcherNil})
		return nil, nil, fmt.Errorf("%s: %w", StrErrUpdateIntegration, ErrFetcherNil)
	}

	data, apiResponse, err := fetcher.Do(ctx, updateReq)
	if err != nil {
		tflog.Error(ctx, StrErrUpdateIntegration, map[string]any{"error": err})
		return nil, apiResponse, fmt.Errorf("%s: %w", StrErrUpdateIntegration, err)
	}
	return data, apiResponse, nil
}
