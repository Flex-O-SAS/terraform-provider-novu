package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	novugo "github.com/novuhq/novu-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EnvironmentsWithKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &EnvironmentsWithKeysDataSource{}
)

func NewEnvironmentsWithKeysDataSource() datasource.DataSource {
	return &EnvironmentsWithKeysDataSource{}
}

// ExampleDataSource defines the data source implementation.
type EnvironmentsWithKeysDataSource struct {
	client *novugo.Novu
}

// EnvironmentsDataSourceModel describes the data source data model.
type EnvironmentsWithKeysDataSourceModel struct {
	Items []EnvironmentWithKeysDataSourceModel `tfsdk:"items"`
}

func (d *EnvironmentsWithKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments_with_keys"
}

func (d *EnvironmentsWithKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environments with API keys data source. Lists all the environments within the organization.",

		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "Environments list",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The environment id",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The environment name",
							Computed:            true,
						},
						// "organization_id": schema.StringAttribute{
						// 	MarkdownDescription: "The environment's organization id",
						// 	Computed:            true,
						// },
						"identifier": schema.StringAttribute{
							MarkdownDescription: "The environment identifier",
							Computed:            true,
						},

						"api_keys": schema.ListNestedAttribute{
							MarkdownDescription: "The environment's api keys",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
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
						"parent_environment_id": schema.StringAttribute{
							MarkdownDescription: "The environment's parent id",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "The environment's slug",
							Computed:            true,
						},
						"is_production": schema.BoolAttribute{
							MarkdownDescription: "Whether it is the production environment",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *EnvironmentsWithKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentsWithKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EnvironmentsWithKeysDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environments, err := d.client.Environments.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environments: %s", err))
		return
	}
	environmentsList := environments.EnvironmentResponseDtos

	items := make([]EnvironmentWithKeysDataSourceModel, 0)
	for _, environment := range environmentsList {
		apiKeys := make([]ApiKeyDataSourceModel, 0)
		for _, apiKey := range environment.APIKeys {
			apiKeys = append(apiKeys, ApiKeyDataSourceModel{
				Value:   types.StringValue(apiKey.Key),
				OwnerId: types.StringValue(apiKey.UserID),
				Hash:    types.StringValue(*apiKey.Hash),
			})
		}

		var isProduction types.Bool

		// Handle nil pointer fields safely
		var parentId types.String
		if environment.ParentID != nil {
			parentId = types.StringValue(*environment.ParentID)
			isProduction = types.BoolValue(true)
		} else {
			parentId = types.StringNull()
			isProduction = types.BoolValue(false)
		}

		var slug types.String
		if environment.Slug != nil {
			slug = types.StringValue(*environment.Slug)
		} else {
			slug = types.StringNull()
		}

		items = append(items, EnvironmentWithKeysDataSourceModel{
			EnvironmentDataSourceModel: EnvironmentDataSourceModel{
				Id:   types.StringValue(environment.ID),
				Name: types.StringValue(environment.Name),
				// OrganizationId: types.StringValue(environment.OrganizationID), // Useless
				Identifier:   types.StringValue(environment.Identifier),
				ParentId:     parentId,
				Slug:         slug,
				IsProduction: isProduction,
			},
			ApiKeys: apiKeys,
		})
	}
	tflog.Info(ctx, "items length", map[string]interface{}{"items": len(items)})

	state.Items = items

	diags := resp.State.Set(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}
