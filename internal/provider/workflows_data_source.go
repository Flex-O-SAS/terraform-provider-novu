package provider

import (
	"context"
	"fmt"
	"terraform-provider-novu/internal/meta"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	novugo "github.com/novuhq/novu-go"
	"github.com/novuhq/novu-go/models/components"
	"github.com/novuhq/novu-go/models/operations"
)

var (
	_ datasource.DataSource              = &WorkflowsDataSource{}
	_ datasource.DataSourceWithConfigure = &WorkflowsDataSource{}

	workflowStatusEnumsPath = "docs/models/components/workflowstatusenum.md"
)

func NewWorkflowsDataSource() datasource.DataSource {
	return &WorkflowsDataSource{}
}

type WorkflowsDataSource struct {
	client *novugo.Novu
}

type WorkflowsDataSourceModel struct {
	Items  []WorkflowDataSourceModel `tfsdk:"items"`
	Tags   []types.String            `tfsdk:"tags"`
	Status []types.String            `tfsdk:"status"`
	Search types.String              `tfsdk:"search"`
}

type WorkflowDataSourceModel struct {
	Id                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Tags              []types.String `tfsdk:"tags"`
	Status            types.String   `tfsdk:"status"`
	UpdatedAt         types.String   `tfsdk:"updated_at"`
	CreatedAt         types.String   `tfsdk:"created_at"`
	WorkflowId        types.String   `tfsdk:"workflow_id"`
	Slug              types.String   `tfsdk:"slug"`
	Origin            types.String   `tfsdk:"origin"`
	LastTriggeredAt   types.String   `tfsdk:"last_triggered_at"`
	StepTypeOverviews []types.String `tfsdk:"step_type_overviews"`
}

func (d *WorkflowsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows"
}

func (d *WorkflowsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List of workflows. Retrieves all the workflows of the API Key's environment used for the provider configuration.",
		Attributes: map[string]schema.Attribute{
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of tags to filter workflows by",
			},
			"status": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of statuses to filter workflows by",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(workflowStatusValidator{}),
				},
			},
			"search": schema.StringAttribute{
				MarkdownDescription: "Case insensitive string to search for in the workflow_id or name of the workflows",
				Optional:            true,
				Description:         "Case insensitive string to search for in the workflow_id or name of the workflows",
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "The list of workflows",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Unique database identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the workflow",
							Computed:            true,
						},
						"tags": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Tags associated with the workflow",
							Computed:            true,
							Optional:            true,
						},
						"updated_at": schema.StringAttribute{
							MarkdownDescription: "Last updated timestamp",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Creation timestamp",
							Computed:            true,
						},

						"workflow_id": schema.StringAttribute{
							MarkdownDescription: "Workflow identifier",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "Workflow slug",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Status of the workflow",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							MarkdownDescription: "Origin of the workflow",
							Computed:            true,
						},
						"last_triggered_at": schema.StringAttribute{
							MarkdownDescription: "Timestamp of the last workflow trigger",
							Computed:            true,
							Optional:            true,
						},
						"step_type_overviews": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Overview of step types in the workflow",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *WorkflowsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// multiple clients, use the novu client
	if clients, ok := req.ProviderData.(*ProviderClients); ok && clients.Novu != nil {
		d.client = clients.Novu
	} else if client, ok := req.ProviderData.(*novugo.Novu); ok {
		// single client novu, use it
		d.client = client
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *novugo.Novu or *ProviderClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
}

func (d *WorkflowsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WorkflowsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowSearchReq := operations.WorkflowControllerSearchWorkflowsRequest{}

	if len(config.Tags) > 0 {
		for _, tag := range config.Tags {
			if tag.IsUnknown() || tag.IsNull() {
				continue
			}
			workflowSearchReq.Tags = append(workflowSearchReq.Tags, tag.ValueString())
		}
	}

	if len(config.Status) > 0 {
		for _, status := range config.Status {
			if status.IsUnknown() || status.IsNull() {
				continue
			}
			workflowSearchReq.Status = append(workflowSearchReq.Status, components.WorkflowStatusEnum(status.ValueString()))
		}
	}

	if !config.Search.IsNull() && !config.Search.IsUnknown() {
		query := config.Search.ValueString()
		workflowSearchReq.Query = &query
	}

	worfklows, err := d.client.Workflows.List(ctx, workflowSearchReq)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list workflows: %v", err))
		return
	}

	workflowsList := worfklows.ListWorkflowResponse.GetWorkflows()

	workflowItems := make([]WorkflowDataSourceModel, 0)
	for _, workflow := range workflowsList {
		var lastTriggeredAt types.String
		if workflow.LastTriggeredAt != nil {
			lastTriggeredAt = types.StringValue(*workflow.LastTriggeredAt)
		} else {
			lastTriggeredAt = types.StringNull()
		}
		stepTypeOverviews := make([]types.String, 0)
		for _, stepType := range workflow.StepTypeOverviews {
			stepTypeOverviews = append(stepTypeOverviews, types.StringValue(string(stepType)))
		}
		tags := make([]types.String, 0)
		for _, tag := range workflow.Tags {
			tags = append(tags, types.StringValue(tag))
		}
		workflowItems = append(workflowItems, WorkflowDataSourceModel{
			Id:                types.StringValue(workflow.ID),
			Name:              types.StringValue(workflow.Name),
			Status:            types.StringValue(string(workflow.Status)),
			UpdatedAt:         types.StringValue(workflow.UpdatedAt),
			CreatedAt:         types.StringValue(workflow.CreatedAt),
			WorkflowId:        types.StringValue(workflow.WorkflowID),
			Slug:              types.StringValue(workflow.Slug),
			Origin:            types.StringValue(string(workflow.Origin)),
			LastTriggeredAt:   lastTriggeredAt,
			StepTypeOverviews: stepTypeOverviews,
			Tags:              tags,
		})
	}
	config.Items = workflowItems

	diags := resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}

type workflowStatusValidator struct{}

func (v workflowStatusValidator) Description(ctx context.Context) string {
	return "Status must be one of " + meta.NovuGoDocsUrl(workflowStatusEnumsPath, "")
}
func (v workflowStatusValidator) MarkdownDescription(ctx context.Context) string {
	return "Status must be one of " + meta.NovuGoDocsUrl(workflowStatusEnumsPath, "")
}

func (v workflowStatusValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	strbt := []byte(req.ConfigValue.ValueString())
	var enum *components.WorkflowStatusEnum
	err := enum.UnmarshalJSON(strbt)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid workflow status",
			"Status must be one of "+meta.NovuGoDocsUrl(workflowStatusEnumsPath, ""),
		)
	}

}
