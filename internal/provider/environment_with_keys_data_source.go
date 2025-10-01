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
	_ datasource.DataSource              = &EnvironmentWithKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &EnvironmentWithKeysDataSource{}
)

type EnvironmentWithKeysDataSourceModel struct {
	ApiKeys []ApiKeyDataSourceModel `tfsdk:"api_keys"`
	EnvironmentDataSourceModel
}

func NewEnvironmentWithKeysDataSource() datasource.DataSource {
	return &EnvironmentWithKeysDataSource{}
}

// ExampleDataSource defines the data source implementation.
type EnvironmentWithKeysDataSource struct {
	client *novugo.Novu
}

func (d *EnvironmentWithKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_with_keys"
}

func (d *EnvironmentWithKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environment data source. Lists the environment with its API keys.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The environment id",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The environment name",
				Optional:            true,
			},
			// "organization_id": schema.StringAttribute{
			// 	MarkdownDescription: "The environment's organization id",
			// 	Computed:            true,
			// },
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The environment identifier",
				Optional:            true,
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
				Optional:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The environment's slug",
				Optional:            true,
			},
			"is_production": schema.BoolAttribute{
				MarkdownDescription: "Whether it is the production environment",
				Optional:            true,
			},
		},
	}
}

func (d *EnvironmentWithKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// multiple clients, use the novu client
	if clients, ok := req.ProviderData.(*ProviderClients); ok && clients.Novu != nil {
		d.client = clients.Novu
	} else if client, ok := req.ProviderData.(*novugo.Novu); ok {
		d.client = client
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *novugo.Novu or *ProviderClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
}

func (d *EnvironmentWithKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EnvironmentWithKeysDataSourceModel
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
	//tflog.Info(ctx, "Environment", map[string]interface{}{"environment": string(jsonData)})
	//tflog.Info(ctx, "Environments", map[string]interface{}{"environments": environmentsList})
	//tflog.Info(ctx, "Environment", map[string]interface{}{"environment": string(jsonData)})

	foundEnvironments := []EnvironmentWithKeysDataSourceModel{}

	for _, environment := range *environmentsList {
		if !state.Id.IsNull() && state.Id.ValueString() != environment.ID {
			tflog.Info(ctx, "Environment id does not match", map[string]interface{}{"environment": environment.ID, "searched": state.Id.ValueString()})
			continue
		}
		if !state.Name.IsNull() && state.Name.ValueString() != environment.Name {
			tflog.Info(ctx, "Environment name does not match", map[string]interface{}{"environment": environment.Name, "searched": state.Name.ValueString()})
			continue
		}
		if !state.Identifier.IsNull() && state.Identifier.ValueString() != environment.Identifier {
			tflog.Info(ctx, "Environment identifier does not match", map[string]interface{}{"environment": environment.Identifier, "searched": state.Identifier.ValueString()})
			continue
		}
		if !state.ParentId.IsNull() && state.ParentId.ValueString() != *environment.ParentID {
			tflog.Info(ctx, "Environment parent id does not match", map[string]interface{}{"environment": *environment.ParentID, "searched": state.ParentId.ValueString()})
			continue
		}
		if !state.Slug.IsNull() && state.Slug.ValueString() != *environment.Slug {
			tflog.Info(ctx, "Environment slug does not match", map[string]interface{}{"environment": *environment.Slug, "searched": state.Slug.ValueString()})
			continue
		}

		isProduction := false
		if environment.ParentID != nil {
			isProduction = true
		}

		if !state.IsProduction.IsNull() && state.IsProduction.ValueBool() != isProduction {
			tflog.Info(ctx, "Environment is_production does not match", map[string]interface{}{"environment": isProduction, "searched": state.IsProduction.ValueBool()})
			continue
		}

		// if we're here, we found an environment that matches the state

		apiKeys := make([]ApiKeyDataSourceModel, 0)
		for _, apiKey := range environment.APIKeys {
			apiKeys = append(apiKeys, ApiKeyDataSourceModel{
				Value:   types.StringValue(apiKey.Key),
				OwnerId: types.StringValue(apiKey.UserID),
				Hash:    types.StringValue(*apiKey.Hash),
			})
		}
		parentId := types.StringNull()
		if environment.ParentID != nil {
			parentId = types.StringValue(*environment.ParentID)
		}
		slug := types.StringNull()
		if environment.Slug != nil {
			slug = types.StringValue(*environment.Slug)
		}
		foundEnvironments = append(foundEnvironments, EnvironmentWithKeysDataSourceModel{
			EnvironmentDataSourceModel: EnvironmentDataSourceModel{
				Id:   types.StringValue(environment.ID),
				Name: types.StringValue(environment.Name),
				// OrganizationId: types.StringValue(environment.OrganizationID), // Useless
				Identifier:   types.StringValue(environment.Identifier),
				ParentId:     parentId,
				Slug:         slug,
				IsProduction: types.BoolValue(isProduction),
			},
			ApiKeys: apiKeys,
		})
	}
	tflog.Info(ctx, "Environments found: ", map[string]interface{}{"environments": len(foundEnvironments)})

	if len(foundEnvironments) == 0 {
		resp.Diagnostics.AddError("No environment found", "No environment matches the criteria")
		return
	}
	if len(foundEnvironments) > 1 {
		names := ""
		for _, environment := range foundEnvironments {
			names += environment.Name.ValueString() + "\n"
		}
		resp.Diagnostics.AddError("Multiple environments found", "Multiple environments found: "+
			"\n\n"+names+
			"\n\n Please use a more specific criteria to find the environment you want to use")
		return
	}

	diags := resp.State.Set(ctx, &foundEnvironments[0])
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}
