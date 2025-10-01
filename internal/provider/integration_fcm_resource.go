package provider

import (
	"context"
	"encoding/json"
	"fmt"
	api_client "terraform-provider-novu/internal/api-client"
	"terraform-provider-novu/internal/helpers"
	customvalidators "terraform-provider-novu/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	novugo "github.com/novuhq/novu-go"
	"github.com/novuhq/novu-go/models/components"
)

const (
	fcmChannel    = components.ChannelPush
	fcmProviderId = components.ProviderIDFcm
)

var (
	_ resource.Resource              = &FCMIntegrationResource{}
	_ resource.ResourceWithConfigure = &FCMIntegrationResource{}
)

type FCMIntegrationResource struct {
	client    *novugo.Novu
	apiClient *api_client.ApiClient
}

type FCMIntegrationResourceModel struct {
	Name       types.String `tfsdk:"name"`
	Identifier types.String `tfsdk:"identifier"`
	// Credentials       types.Object            `tfsdk:"credentials"`
	Active            types.Bool              `tfsdk:"active"`
	Check             types.Bool              `tfsdk:"check"`
	Conditions        types.List              `tfsdk:"conditions"`
	JsonConfiguration types.String            `tfsdk:"json_configuration"`
	ServiceAccount    *FcmServiceAccountModel `tfsdk:"service_account"`
}

type FcmServiceAccountModel struct {
	Type                    types.String `tfsdk:"type"`
	ProjectId               types.String `tfsdk:"project_id"`
	PrivateKeyId            types.String `tfsdk:"private_key_id"`
	PrivateKey              types.String `tfsdk:"private_key"`
	ClientEmail             types.String `tfsdk:"client_email"`
	ClientId                types.String `tfsdk:"client_id"`
	AuthUri                 types.String `tfsdk:"auth_uri"`
	TokenUri                types.String `tfsdk:"token_uri"`
	AuthProviderX509CertUrl types.String `tfsdk:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       types.String `tfsdk:"client_x509_cert_url"`
}

type fcmServiceAccount struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}

func NewFCMIntegrationResource() resource.Resource {
	return &FCMIntegrationResource{}
}

func (r *FCMIntegrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fcm_integration"
}

func (r *FCMIntegrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if clients, ok := req.ProviderData.(*ProviderClients); ok && clients.Novu != nil && clients.Api != nil {
		r.client = clients.Novu
		r.apiClient = clients.Api
	} else {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *ProviderClients or *novugo.Novu, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
}

func (r *FCMIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "FCM Integration",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the integration",
				Optional:            true,
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "Identifier of the integration",
				Optional:            true,
			},
			// "credentials": schema.SingleNestedAttribute{
			// 	MarkdownDescription: "Credentials of the integration",
			// 	Optional:            true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"service_account": schema.StringAttribute{
			// 			MarkdownDescription: "Service account of the integration",
			// 			Optional:            true,
			// 		},
			// 	},
			// },
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the integration is active",
				Optional:            true,
			},
			"check": schema.BoolAttribute{
				MarkdownDescription: "Whether the integration is checked",
				Optional:            true,
			},
			"conditions": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Conditions of the integration",
				Optional:            true,
			},
			"json_configuration": schema.StringAttribute{
				MarkdownDescription: "JSON configuration of the integration",
				Optional:            true,
				Validators: []validator.String{
					customvalidators.ExactlyOneSet("service_account"),
				},
			},
			"service_account": schema.SingleNestedAttribute{
				MarkdownDescription: "FCM service account",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of the service account",
						Required:            true,
					},
					"project_id": schema.StringAttribute{
						MarkdownDescription: "Project ID of the service account",
						Required:            true,
					},
					"private_key_id": schema.StringAttribute{
						MarkdownDescription: "Private key ID of the service account",
						Required:            true,
					},
					"private_key": schema.StringAttribute{
						MarkdownDescription: "Private key of the service account",
						Required:            true,
					},
					"client_email": schema.StringAttribute{
						MarkdownDescription: "Client email of the service account",
						Required:            true,
					},
					"client_id": schema.StringAttribute{
						MarkdownDescription: "Client ID of the service account",
						Required:            true,
					},
					"auth_uri": schema.StringAttribute{
						MarkdownDescription: "Auth URI of the service account",
						Required:            true,
					},
					"token_uri": schema.StringAttribute{
						MarkdownDescription: "Token URI of the service account",
						Required:            true,
					},
					"auth_provider_x509_cert_url": schema.StringAttribute{
						MarkdownDescription: "Auth provider X509 cert URL of the service account",
						Required:            true,
					},
					"client_x509_cert_url": schema.StringAttribute{
						MarkdownDescription: "Client X509 cert URL of the service account",
						Required:            true,
					},
				},
			},
		},
	}
}

func (r *FCMIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FCMIntegrationResourceModel
	var state FCMIntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the integration already exists (the SDK does not properly handle this)
	var foundIntegration components.IntegrationResponseDto
	integration, err := r.client.Integrations.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get integration: %s", err))
		return
	}
	integrationList := integration.IntegrationResponseDtos
	for _, integration := range integrationList {
		if integration.Identifier == state.Identifier.ValueString() {
			foundIntegration = integration
			break
		}
	}
	if foundIntegration.Identifier != "" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Integration with identifier %s already exists", state.Identifier.ValueString()))
		return
	}

	createReq := components.CreateIntegrationRequestDto{
		Channel:    components.CreateIntegrationRequestDtoChannel(fcmChannel),
		ProviderID: string(fcmProviderId),
		Name:       helpers.FromTfString(plan.Name),
		Identifier: helpers.FromTfString(plan.Identifier),
		Active:     helpers.FromTfBool(plan.Active),
		Check:      helpers.FromTfBool(plan.Check),
		Conditions: []components.StepFilterDto{},
	}

	if !plan.JsonConfiguration.IsNull() && !plan.JsonConfiguration.IsUnknown() {
		credentials := components.CredentialsDto{
			ServiceAccount: helpers.FromTfString(plan.JsonConfiguration),
		}
		createReq.Credentials = &credentials
	}

	res, apiResponse, err := r.apiClient.CreateIntegration(ctx, &createReq)
	if err != nil {
		if apiResponse != nil && apiResponse.StatusCode == 429 {
			resp.Diagnostics.AddError("Client Error", "Unable to create integration: Integration already exists")
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create integration: %s", err))
		return
	}
	if res == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create integration: Response is nil")
		return
	}

	resp.Diagnostics.Append(state.setFromResponseDTO(ctx, &plan, res)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Client Error", "Unable to set response to data")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// TODO
func (r *FCMIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// TODO
func (r *FCMIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FCMIntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Identifier.IsNull() || state.Identifier.IsUnknown() || state.Identifier.ValueString() == "" {
		resp.Diagnostics.AddError("Client Error", "identifier is required but was not set")
		return
	}

	var foundIntegration components.IntegrationResponseDto
	integration, err := r.client.Integrations.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get integration: %s", err))
		return
	}
	integrationList := integration.IntegrationResponseDtos
	for _, integration := range integrationList {
		if integration.Identifier == state.Identifier.ValueString() {
			foundIntegration = integration
			break
		}
	}

	if foundIntegration.Identifier == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(state.setFromResponseDTO(ctx, &state, &foundIntegration)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// TODO
func (r *FCMIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// TODO
func (r *FCMIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	return
}

func (data *FCMIntegrationResourceModel) setFromResponseDTO(_ context.Context, previous *FCMIntegrationResourceModel, integration *components.IntegrationResponseDto) diag.Diagnostics {
	var rdiags diag.Diagnostics

	if integration == nil {
		return rdiags
	}

	if data == nil {
		data = &FCMIntegrationResourceModel{}
	}

	if previous == nil {
		previous = &FCMIntegrationResourceModel{}
	}

	data.Check = previous.Check

	data.Name = types.StringValue(integration.Name)
	data.Identifier = types.StringValue(integration.Identifier)
	data.Active = types.BoolValue(integration.Active)
	data.Conditions = types.ListNull(types.StringType)

	if integration.Credentials.ServiceAccount != nil {
		if !previous.JsonConfiguration.IsNull() && !previous.JsonConfiguration.IsUnknown() {
			// we fill JsonConfiguration
			data.JsonConfiguration = helpers.TfString(integration.Credentials.ServiceAccount)

		} else {
			// we fill service_account
			fcmServiceAccount := fcmServiceAccount{}
			err := json.Unmarshal([]byte(*integration.Credentials.ServiceAccount), &fcmServiceAccount)
			if err != nil {
				rdiags.AddError("Client Error", fmt.Sprintf("Error unmarshalling FCM service account: %s", err))
				return rdiags
			}
			data.ServiceAccount = &FcmServiceAccountModel{
				Type:                    types.StringValue(fcmServiceAccount.Type),
				ProjectId:               types.StringValue(fcmServiceAccount.ProjectId),
				PrivateKeyId:            types.StringValue(fcmServiceAccount.PrivateKeyId),
				PrivateKey:              types.StringValue(fcmServiceAccount.PrivateKey),
				ClientEmail:             types.StringValue(fcmServiceAccount.ClientEmail),
				ClientId:                types.StringValue(fcmServiceAccount.ClientId),
				AuthUri:                 types.StringValue(fcmServiceAccount.AuthUri),
				TokenUri:                types.StringValue(fcmServiceAccount.TokenUri),
				AuthProviderX509CertUrl: types.StringValue(fcmServiceAccount.AuthProviderX509CertUrl),
				ClientX509CertUrl:       types.StringValue(fcmServiceAccount.ClientX509CertUrl),
			}
		}
	}

	return nil
}
