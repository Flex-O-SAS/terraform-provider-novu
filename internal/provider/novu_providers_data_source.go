package provider

import (
	"context"
	"terraform-provider-novu/internal/meta"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/novuhq/novu-go/models/components"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProvidersDataSource{}

func NewProvidersDataSource() datasource.DataSource {
	return &ProvidersDataSource{}
}

// ProvidersDataSource defines the data source implementation.
type ProvidersDataSource struct {
}

type ProvidersDataSourceModel struct {
	Items []types.String `tfsdk:"items"`
}

func (d *ProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_providers"
}

func (d *ProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Non exhaustive list of supported novu providers. The actual supported list is available in the novu go SDK doc: " + meta.NovuGoDocsUrl(providersEnumsPath, "main") +
			"\n If your provider is not in this list but is present in " + meta.NovuGoDocsUrl(providersEnumsPath, "main") + ", the terraform provider will not be able to use it until it is updated.",

		Attributes: map[string]schema.Attribute{
			"items": schema.ListAttribute{
				MarkdownDescription: "List of supported providers",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}

}

func (d *ProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProvidersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var allProviders = []components.ProvidersIDEnum{
		components.ProvidersIDEnumEmailjs,
		components.ProvidersIDEnumMailgun,
		components.ProvidersIDEnumMailjet,
		components.ProvidersIDEnumMandrill,
		components.ProvidersIDEnumNodemailer,
		components.ProvidersIDEnumPostmark,
		components.ProvidersIDEnumSendgrid,
		components.ProvidersIDEnumSendinblue,
		components.ProvidersIDEnumSes,
		components.ProvidersIDEnumNetcore,
		components.ProvidersIDEnumInfobipEmail,
		components.ProvidersIDEnumResend,
		components.ProvidersIDEnumPlunk,
		components.ProvidersIDEnumMailersend,
		components.ProvidersIDEnumMailtrap,
		components.ProvidersIDEnumClickatell,
		components.ProvidersIDEnumOutlook365,
		components.ProvidersIDEnumNovuEmail,
		components.ProvidersIDEnumSparkpost,
		components.ProvidersIDEnumEmailWebhook,
		components.ProvidersIDEnumBraze,
		components.ProvidersIDEnumNexmo,
		components.ProvidersIDEnumPlivo,
		components.ProvidersIDEnumSms77,
		components.ProvidersIDEnumSmsCentral,
		components.ProvidersIDEnumSns,
		components.ProvidersIDEnumTelnyx,
		components.ProvidersIDEnumTwilio,
		components.ProvidersIDEnumGupshup,
		components.ProvidersIDEnumFiretext,
		components.ProvidersIDEnumInfobipSms,
		components.ProvidersIDEnumBurstSms,
		components.ProvidersIDEnumBulkSms,
		components.ProvidersIDEnumIsendSms,
		components.ProvidersIDEnumFortySixElks,
		components.ProvidersIDEnumKannel,
		components.ProvidersIDEnumMaqsam,
		components.ProvidersIDEnumTermii,
		components.ProvidersIDEnumAfricasTalking,
		components.ProvidersIDEnumNovuSms,
		components.ProvidersIDEnumSendchamp,
		components.ProvidersIDEnumGenericSms,
		components.ProvidersIDEnumClicksend,
		components.ProvidersIDEnumBandwidth,
		components.ProvidersIDEnumMessagebird,
		components.ProvidersIDEnumSimpletexting,
		components.ProvidersIDEnumAzureSms,
		components.ProvidersIDEnumRingCentral,
		components.ProvidersIDEnumBrevoSms,
		components.ProvidersIDEnumEazySms,
		components.ProvidersIDEnumMobishastra,
		components.ProvidersIDEnumAfroMessage,
		components.ProvidersIDEnumFcm,
		components.ProvidersIDEnumApns,
		components.ProvidersIDEnumExpo,
		components.ProvidersIDEnumOneSignal,
		components.ProvidersIDEnumPushpad,
		components.ProvidersIDEnumPushWebhook,
		components.ProvidersIDEnumPusherBeams,
		components.ProvidersIDEnumNovu,
		components.ProvidersIDEnumSlack,
		components.ProvidersIDEnumDiscord,
		components.ProvidersIDEnumMsteams,
		components.ProvidersIDEnumMattermost,
		components.ProvidersIDEnumRyver,
		components.ProvidersIDEnumZulip,
		components.ProvidersIDEnumGrafanaOnCall,
		components.ProvidersIDEnumGetstream,
		components.ProvidersIDEnumRocketChat,
		components.ProvidersIDEnumWhatsappBusiness,
	}

	data.Items = make([]types.String, 0)
	for _, provider := range allProviders {
		data.Items = append(data.Items, types.StringValue(string(provider)))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
