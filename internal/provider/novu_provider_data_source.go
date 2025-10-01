package provider

import (
	"context"
	"fmt"
	"terraform-provider-novu/internal/meta"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/novuhq/novu-go/models/components"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource = &ProviderDataSource{}

	providersEnumsPath = "docs/models/components/providersidenum.md"
)

func NewProviderDataSource() datasource.DataSource {
	return &ProviderDataSource{}
}

// ProviderDataSource defines the data source implementation.
type ProviderDataSource struct {
}

// ProviderDataSourceModel describes the data source data model.
type ProviderDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

func (d *ProviderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (d *ProviderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provider data source. The available providers are listed in the novu go SDK, and are subject to change. This data source only checks if the provider exists in the current version of the novu go SDK, it does not make any API call as Novu does not have a dedicated API for providers." +
			"\n You may find the latest list of supported providers in the novu go SDK doc: " + meta.NovuGoDocsUrl(providersEnumsPath, "") +
			"\n If your provider is not in this list but is present in " + meta.NovuGoDocsUrl(providersEnumsPath, "main") + ", the terraform provider will not be able to use it until it is updated.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Provider name",
				Required:            true,
			},
		},
	}
}

func (d *ProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var provider components.ProvidersIDEnum
	tflog.Info(ctx, "Provider name", map[string]interface{}{"name": data.Name.ValueString()})
	providerJSON := fmt.Sprintf(`"%s"`, data.Name.ValueString())
	err := provider.UnmarshalJSON([]byte(providerJSON))
	if err != nil {
		resp.Diagnostics.AddError("Provider Error", fmt.Sprintf("Either the novu provider %s does not exist, or it has been recently added and the terraform provider must be updated."+
			"\n You may find the latest list of supported providers in the novu go SDK doc: "+meta.NovuGoDocsUrl(providersEnumsPath, "")+
			"\n If your provider is not in this list but is present in "+meta.NovuGoDocsUrl(providersEnumsPath, "main")+", the terraform provider will not be able to use it until it is updated."+
			"\n NB : This source only checks if the provider exists in the current version of the novu go SDK, it does not make any API call as Novu does not have a dedicated API for providers.",
			data.Name.ValueString()))
		return
	}

	data.Name = types.StringValue(string(provider))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
