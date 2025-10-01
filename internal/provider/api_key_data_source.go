package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	novugo "github.com/novuhq/novu-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ApiKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &ApiKeyDataSource{}
)

func NewApiKeyDataSource() datasource.DataSource {
	return &ApiKeyDataSource{}
}

type ApiKeyDataSource struct {
	client *novugo.Novu
}

type ApiKeyDataSourceModel struct {
	Value   types.String `tfsdk:"value"`
	OwnerId types.String `tfsdk:"owner_id"`
	Hash    types.String `tfsdk:"hash"`
}

type EnvironmentApiKeyDataSourceModel struct {
	EnvironmentId types.String `tfsdk:"environment_id"`
	Value         types.String `tfsdk:"value"`
	OwnerId       types.String `tfsdk:"owner_id"`
	Hash          types.String `tfsdk:"hash"`
}

func (d *ApiKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *ApiKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Api key data source. Retrieve a single API key for an environment if set, otherwise try to retrieve a single API key for all the environments within the organization. Will fail if multiple API keys are found.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The environment id",
				Optional:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The api key value",
				Computed:            true,
				Sensitive:           true,
			},
			"owner_id": schema.StringAttribute{
				MarkdownDescription: "ID of the user that owns the api key",
				Computed:            true,
			},
			"hash": schema.StringAttribute{
				MarkdownDescription: "The api key hash",
				Computed:            true,
			},
		},
	}
}

func (d *ApiKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EnvironmentApiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environments, err := d.client.Environments.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environments, got error: %s", err))
		return
	}
	environmentsList := &environments.EnvironmentResponseDtos

	apiKeys := make([]EnvironmentApiKeyDataSourceModel, 0)

	for _, environment := range *environmentsList {
		if !state.EnvironmentId.IsNull() && state.EnvironmentId.ValueString() != environment.ID {
			continue
		}
		for _, apiKey := range environment.APIKeys {
			apiKeys = append(apiKeys, EnvironmentApiKeyDataSourceModel{
				EnvironmentId: types.StringValue(environment.ID),
				Value:         types.StringValue(apiKey.Key),
				OwnerId:       types.StringValue(apiKey.UserID),
				Hash:          types.StringValue(*apiKey.Hash),
			})
		}
	}

	if len(apiKeys) == 0 {
		envString := "in any environment"
		if !state.EnvironmentId.IsNull() {
			envString = "for environment " + state.EnvironmentId.ValueString()
		}
		resp.Diagnostics.AddError("Client Error", "No api key found "+envString)
		return
	}
	if len(apiKeys) > 1 {
		resp.Diagnostics.AddError("Client Error", "Multiple api keys found")
		return
	}

	state.Value = apiKeys[0].Value
	state.OwnerId = apiKeys[0].OwnerId
	state.Hash = apiKeys[0].Hash
	state.EnvironmentId = apiKeys[0].EnvironmentId

	diags := resp.State.Set(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}
