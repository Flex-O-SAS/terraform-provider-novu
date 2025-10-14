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
	_ datasource.DataSource              = &EnvironmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &EnvironmentsDataSource{}
)

func NewEnvironmentsDataSource() datasource.DataSource {
	return &EnvironmentsDataSource{}
}

// ExampleDataSource defines the data source implementation.
type EnvironmentsDataSource struct {
	client *novugo.Novu
}

// EnvironmentsDataSourceModel describes the data source data model.
type EnvironmentsDataSourceModel struct {
	Items []EnvironmentDataSourceModel `tfsdk:"items"`
}

func (d *EnvironmentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *EnvironmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environments data source. Lists all the environments within the organization.",

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

func (d *EnvironmentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// multiple clients, use the novu client
	if clients, ok := req.ProviderData.(*ProviderClients); ok && clients.Novu != nil {
		d.client = clients.Novu
		// single client novu, use it
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

func (d *EnvironmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EnvironmentsDataSourceModel

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
	//tflog.Info(ctx, "Environment", map[string]interface{}{"environment": string(jsonData)})
	//tflog.Info(ctx, "Environments", map[string]interface{}{"environments": environmentsList})
	//tflog.Info(ctx, "Environment", map[string]interface{}{"environment": string(jsonData)})

	items := make([]EnvironmentDataSourceModel, 0)
	for _, environment := range environmentsList {
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

		items = append(items, EnvironmentDataSourceModel{
			Id:   types.StringValue(environment.ID),
			Name: types.StringValue(environment.Name),
			// OrganizationId: types.StringValue(environment.OrganizationID), // Useless
			Identifier:   types.StringValue(environment.Identifier),
			ParentId:     parentId,
			Slug:         slug,
			IsProduction: isProduction,
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
