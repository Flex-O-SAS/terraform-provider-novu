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
	_ datasource.DataSource              = &ApiKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &ApiKeysDataSource{}
)

func NewApiKeysDataSource() datasource.DataSource {
	return &ApiKeysDataSource{}
}

type ApiKeysDataSource struct {
	client *novugo.Novu
}

type EnvironmentApiKeysDataSourceModel struct {
	EnvironmentId types.String                       `tfsdk:"environment_id"`
	Items         []EnvironmentApiKeyDataSourceModel `tfsdk:"items"`
}

func (d *ApiKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *ApiKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Api keys data source. Retrieve all the API keys for an environment if set, otherwise retrieve all the API keys for all the environments within the organization",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The environment id",
				Optional:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "The environment's api keys",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"environment_id": schema.StringAttribute{
							MarkdownDescription: "The environment id",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The api key",
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
				},
			},
		},
	}
}

func (d *ApiKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApiKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config EnvironmentApiKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.EnvironmentId.IsUnknown() {
		resp.Diagnostics.AddError("Client Error", "Environment id is unknown, impossible to retrieve the api keys")
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
		if !config.EnvironmentId.IsNull() && config.EnvironmentId.ValueString() != environment.ID {
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

	config.Items = apiKeys

	diags := resp.State.Set(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
}
