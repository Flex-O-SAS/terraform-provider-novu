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

// Option 1 : récupère le body en []byte, obligé d'unmarshal à l'intérieur de la fonction
func (c *ApiClient) GetWorkflow(ctx context.Context, workflowID string) (*WorkflowResponseDto, error) {
	clientResponse, err := c.Get(ctx, workflowPath+workflowID, nil)
	if err != nil {
		tflog.Error(ctx, "error getting workflow", map[string]any{"error": err})
		return nil, err
	}

	if clientResponse == nil {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": ErrClientResponseNil})
		return nil, fmt.Errorf("%s: %w", StrErrGetWorkflow, ErrClientResponseNil)
	}

	tflog.Debug(ctx, "clientResponse body", map[string]any{
		"body":       string(clientResponse.Body),
		"status":     clientResponse.StatusCode,
		"statusText": clientResponse.Status,
	})

	// todo: unmarshal body to get more details
	if clientResponse.StatusCode != 200 {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": fmt.Errorf("%w, got: %s", ErrClientExpected200, clientResponse.Status)})
		return nil, fmt.Errorf("%s: %w, got: %s", StrErrGetWorkflow, ErrClientExpected200, clientResponse.Status)
	}

	var workflowResponseDto WorkflowResponseDto
	err = unmarshalBodyKey(clientResponse.Body, &workflowResponseDto, "data")
	// or err = unmarshalBodyData(res.Body, workflowResponseDto)
	if err != nil {
		tflog.Error(ctx, StrErrUnmarshalWorkflow, map[string]any{"error": err})
		return nil, fmt.Errorf("%s: %s", StrErrGetWorkflow, StrErrUnmarshalWorkflow)
	}

	if workflowResponseDto.ID == "" {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": ErrDataEmpty})
		return nil, fmt.Errorf("%s: %s", StrErrGetWorkflow, ErrDataEmpty)
	}
	return &workflowResponseDto, nil
}

// Option 2 : le fetcher déduit le type à renvoyer et s'occupe de l'unmarshal
func (c *ApiClient) GetWorkflowPolymorphic(ctx context.Context, workflowID string) (*WorkflowResponseDto, *ApiClientResponse, error) {
	uri, err := url.JoinPath(workflowPath, workflowID)
	if err != nil {
		tflog.Error(ctx, StrErrGetWorkflow, map[string]any{"error": err})
		return nil, nil, fmt.Errorf("%s: %w", StrErrGetWorkflow, err)
	}

	fetcher := fetcher[WorkflowResponseDto](c, http.MethodGet, uri) // url concat better
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

func (c *ApiClient) CreateWorkflow(ctx context.Context, createWorkflowDto *components.CreateWorkflowDto) (*WorkflowResponseDto, error) {
	fetcher := fetcher[WorkflowResponseDto](c, http.MethodPost, workflowPath)
	if fetcher == nil {
		tflog.Error(ctx, StrErrCreateWorkflow, map[string]any{"error": ErrFetcherNil})
		return nil, fmt.Errorf("%s: %w", StrErrCreateWorkflow, ErrFetcherNil)
	}

	data, _, err := fetcher.Do(ctx, createWorkflowDto)
	if err != nil {
		tflog.Error(ctx, StrErrCreateWorkflow, map[string]any{"error": err})
		return nil, fmt.Errorf("%s: %w", StrErrCreateWorkflow, err)
	}

	return data, nil
}

func (c *ApiClient) UpdateWorkflow(ctx context.Context, workflowID string, updateReq *components.UpdateWorkflowDto) (*WorkflowResponseDto, error) {
	uri, err := url.JoinPath(workflowPath, workflowID)
	if err != nil {
		tflog.Error(ctx, StrErrUpdateWorkflow, map[string]any{"error": err})
		return nil, fmt.Errorf("%s: %w", StrErrUpdateWorkflow, err)
	}

	fetcher := fetcher[WorkflowResponseDto](c, http.MethodPut, uri)
	if fetcher == nil {
		tflog.Error(ctx, StrErrUpdateWorkflow, map[string]any{"error": ErrFetcherNil})
		return nil, fmt.Errorf("%s: %w", StrErrUpdateWorkflow, ErrFetcherNil)
	}

	data, _, err := fetcher.Do(ctx, updateReq)
	if err != nil {
		tflog.Error(ctx, StrErrUpdateWorkflow, map[string]any{"error": err})
		return nil, fmt.Errorf("%s: %w", StrErrUpdateWorkflow, err)
	}

	return data, nil
}
