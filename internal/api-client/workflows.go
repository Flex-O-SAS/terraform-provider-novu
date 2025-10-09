package api_client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/novuhq/novu-go/models/components"
)

const (
	StrErrGetWorkflow       = "error getting workflow"
	StrErrCreateWorkflow    = "error creating workflow"
	StrErrUnmarshalWorkflow = "error unmarshalling workflow"
	StrErrUpdateWorkflow    = "error updating workflow"

	workflowPath = "v2/workflows/"
)

// Needed because the SDK cannot return a 404 status code
func (c *ApiClient) GetWorkflow(ctx context.Context, workflowID string) (*components.WorkflowResponseDto, *ApiClientResponse, error) {
	uri, err := url.JoinPath(workflowPath, workflowID)
	if err != nil {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": err})
		return nil, nil, fmt.Errorf("%s: %w", StrErrGetWorkflow, err)
	}

	fetcher := fetcher[components.WorkflowResponseDto](c, http.MethodGet, uri) // url concat better
	// peu probable que fetcher soit nil
	if fetcher == nil {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": ErrFetcherNil})
		return nil, nil, fmt.Errorf("%s: %w", StrErrGetWorkflow, ErrFetcherNil)
	}

	data, clientResponse, err := fetcher.Do(ctx, nil)

	if err != nil {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": err})
		return nil, clientResponse, fmt.Errorf("%s: %w", StrErrGetWorkflow, err)
	}

	if clientResponse == nil {
		//fixme : do I actually need to log errors at this level ? Also, do I need to add strErrGetWorkflow ?
		// Same question for the rest
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": ErrClientResponseNil})
		return nil, clientResponse, fmt.Errorf("%s: %w", StrErrGetWorkflow, ErrClientResponseNil)
	}

	tflog.Debug(ctx, "clientResponse body", map[string]any{
		"body":       string(clientResponse.Body),
		"status":     clientResponse.StatusCode,
		"statusText": clientResponse.Status,
	})

	// todo: get more details from the body
	if clientResponse.StatusCode != 200 {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": fmt.Errorf("%w, got: %s", ErrClientExpected200, clientResponse.Status)})
		return nil, clientResponse, fmt.Errorf("%s: %w, got: %s", StrErrGetWorkflow, ErrClientExpected200, clientResponse.Status)
	}

	if data == nil {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": ErrDataNil})
		return nil, clientResponse, fmt.Errorf("%s: %w", StrErrGetWorkflow, ErrDataNil)
	}

	if data.ID == "" {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": ErrDataEmpty})
		return nil, clientResponse, fmt.Errorf("%s: %s", StrErrGetWorkflow, ErrDataEmpty)
	}

	return data, clientResponse, nil
}
