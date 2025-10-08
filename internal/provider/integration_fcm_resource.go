package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	api_client "terraform-provider-novu/internal/api-client"
	"terraform-provider-novu/internal/helpers"
	customvalidators "terraform-provider-novu/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	ID                types.String            `tfsdk:"id"`
	Name              types.String            `tfsdk:"name"`
	Identifier        types.String            `tfsdk:"identifier"`
	Active            types.Bool              `tfsdk:"active"`
	Check             types.Bool              `tfsdk:"check"`
	JsonConfiguration types.String            `tfsdk:"json_configuration"`
	ServiceAccount    *FcmServiceAccountModel `tfsdk:"service_account"`
	// Credentials       types.Object            `tfsdk:"credentials"` // TODO : implement or remove ?

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
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the integration",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
				Computed:            true,
				Default:             booldefault.StaticBool(false), // API default
			},
			"check": schema.BoolAttribute{
				MarkdownDescription: "Whether the integration is checked. It is recommended to set this to true.",
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
	} else if plan.ServiceAccount != nil {
		serviceAccountJson, err := plan.ServiceAccount.ToFcmServiceAccountJsonStr()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error marshalling FCM service account: %s", err))
			return
		}
		createReq.Credentials = &components.CredentialsDto{
			ServiceAccount: &serviceAccountJson,
		}
	}

	integrationRes, _, err := r.apiClient.CreateIntegration(ctx, &createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	if integrationRes == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create integration: Response is nil")
		return
	}

	resp.Diagnostics.Append(state.setFromResponseDTO(ctx, &plan, integrationRes)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Client Error", "Unable to set response to data")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FCMIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state FCMIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() || state.ID.ValueString() == "" {
		resp.Diagnostics.AddError("Client Error", "id is required but was not set")
		return
	}

	_, err := r.client.Integrations.Delete(ctx, state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete integration: %s", err))
		return
	}
}

// TODO
func (r *FCMIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FCMIntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	IDSet := !state.ID.IsNull() && !state.ID.IsUnknown() && state.ID.ValueString() != ""
	IdentifierSet := !state.Identifier.IsNull() && !state.Identifier.IsUnknown() && state.Identifier.ValueString() != ""

	stateID := state.ID.ValueString()
	stateIdentifier := state.Identifier.ValueString()

	if !IDSet && !IdentifierSet {
		resp.Diagnostics.AddError("Client Error", "identifier or id is required but was not set")
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
		if IDSet && *integration.ID == stateID {
			foundIntegration = integration
			break
		} else if !IDSet && IdentifierSet && integration.Identifier == stateIdentifier {
			// Only search by identifier if ID is not set (ex: case of import)
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
	var state FCMIntegrationResourceModel
	var plan FCMIntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.IsUnknown() || state.ID.ValueString() == "" {
		resp.Diagnostics.AddError("Client Error", "id is required but was not set")
		return
	}

	var updateReq components.UpdateIntegrationRequestDto
	helpers.SetStringPtrIfChanged(plan.Name, state.Name, &updateReq.Name)
	helpers.SetStringPtrIfChanged(plan.Identifier, state.Identifier, &updateReq.Identifier)
	helpers.SetBoolPtrIfChanged(plan.Active, state.Active, &updateReq.Active)
	helpers.SetBoolPtrIfChanged(plan.Check, state.Check, &updateReq.Check)

	if !plan.JsonConfiguration.Equal(state.JsonConfiguration) || !plan.ServiceAccount.Equal(state.ServiceAccount) {
		if !state.JsonConfiguration.IsNull() && !state.JsonConfiguration.IsUnknown() {
			updateReq.Credentials = &components.CredentialsDto{
				ServiceAccount: helpers.FromTfString(state.JsonConfiguration),
			}
		} else if plan.ServiceAccount != nil {
			serviceAccountJson, err := plan.ServiceAccount.ToFcmServiceAccountJsonStr()
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error marshalling FCM service account: %s", err))
				return
			}
			serviceAccountJsonStr := serviceAccountJson
			updateReq.Credentials = &components.CredentialsDto{
				ServiceAccount: &serviceAccountJsonStr,
			}
		} else {
			updateReq.Credentials = &components.CredentialsDto{}
		}
	}
	if updateReq.Name == nil && updateReq.Identifier == nil &&
		updateReq.Active == nil && updateReq.Check == nil &&
		updateReq.Credentials == nil {
		return
	}

	integrationRes, _, err := r.apiClient.UpdateIntegration(ctx, state.ID.ValueString(), &updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update integration: %s", err))
		return
	}
	if integrationRes == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update integration: Response is nil")
		return
	}

	resp.Diagnostics.Append(state.setFromResponseDTO(ctx, &plan, integrationRes)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Client Error", "Unable to set response to data")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// TODO : do we use ID instead ? it's not directly accessible, either through the list API or by inspecting the integration with the web console
func (r *FCMIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
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

	data.ID = helpers.TfString(integration.ID)
	data.Name = types.StringValue(integration.Name)
	data.Identifier = types.StringValue(integration.Identifier)
	data.Active = types.BoolValue(integration.Active)

	// fixme : remove these two lines then use below code once novu fixed their API
	// NB : the code below could not be tested
	data.ServiceAccount = previous.ServiceAccount
	data.JsonConfiguration = previous.JsonConfiguration

	// TODO : add a json comparator to skip whitespace changes
	// if integration.Credentials.ServiceAccount != nil {
	// 	if !previous.JsonConfiguration.IsNull() && !previous.JsonConfiguration.IsUnknown() {
	// 		// we fill JsonConfiguration
	// 		data.JsonConfiguration = helpers.TfString(integration.Credentials.ServiceAccount)
	// 	} else {
	// 		// we fill service_account
	// 		fcmServiceAccount := fcmServiceAccount{}
	// 		err := json.Unmarshal([]byte(*integration.Credentials.ServiceAccount), &fcmServiceAccount)
	// 		if err != nil {
	// 			rdiags.AddError("Client Error", fmt.Sprintf("Error unmarshalling FCM service account: %s", err))
	// 			return rdiags
	// 		}
	// 		data.ServiceAccount = &FcmServiceAccountModel{
	// 			Type:                    types.StringValue(fcmServiceAccount.Type),
	// 			ProjectId:               types.StringValue(fcmServiceAccount.ProjectId),
	// 			PrivateKeyId:            types.StringValue(fcmServiceAccount.PrivateKeyId),
	// 			PrivateKey:              types.StringValue(fcmServiceAccount.PrivateKey),
	// 			ClientEmail:             types.StringValue(fcmServiceAccount.ClientEmail),
	// 			ClientId:                types.StringValue(fcmServiceAccount.ClientId),
	// 			AuthUri:                 types.StringValue(fcmServiceAccount.AuthUri),
	// 			TokenUri:                types.StringValue(fcmServiceAccount.TokenUri),
	// 			AuthProviderX509CertUrl: types.StringValue(fcmServiceAccount.AuthProviderX509CertUrl),
	// 			ClientX509CertUrl:       types.StringValue(fcmServiceAccount.ClientX509CertUrl),
	// 		}
	// 	}
	// }

	return nil
}

func (data *FcmServiceAccountModel) Equal(other *FcmServiceAccountModel) bool {
	if data == nil && other == nil {
		return true
	}
	if data == nil || other == nil {
		return false
	}
	return data.Type.Equal(other.Type) &&
		data.ProjectId.Equal(other.ProjectId) &&
		data.PrivateKeyId.Equal(other.PrivateKeyId) &&
		data.PrivateKey.Equal(other.PrivateKey) &&
		data.ClientEmail.Equal(other.ClientEmail) &&
		data.ClientId.Equal(other.ClientId) &&
		data.AuthUri.Equal(other.AuthUri) &&
		data.TokenUri.Equal(other.TokenUri) &&
		data.AuthProviderX509CertUrl.Equal(other.AuthProviderX509CertUrl) &&
		data.ClientX509CertUrl.Equal(other.ClientX509CertUrl)
}

func (data *FcmServiceAccountModel) ToFcmServiceAccount() (fcmServiceAccount, bool) {
	hasUnknown := data.Type.IsUnknown() ||
		data.ProjectId.IsUnknown() ||
		data.PrivateKeyId.IsUnknown() ||
		data.PrivateKey.IsUnknown() ||
		data.ClientEmail.IsUnknown() ||
		data.ClientId.IsUnknown() || data.AuthUri.IsUnknown() ||
		data.TokenUri.IsUnknown() ||
		data.AuthProviderX509CertUrl.IsUnknown() ||
		data.ClientX509CertUrl.IsUnknown()

	return fcmServiceAccount{
		Type:                    data.Type.ValueString(),
		ProjectId:               data.ProjectId.ValueString(),
		PrivateKeyId:            data.PrivateKeyId.ValueString(),
		PrivateKey:              data.PrivateKey.ValueString(),
		ClientEmail:             data.ClientEmail.ValueString(),
		ClientId:                data.ClientId.ValueString(),
		AuthUri:                 data.AuthUri.ValueString(),
		TokenUri:                data.TokenUri.ValueString(),
		AuthProviderX509CertUrl: data.AuthProviderX509CertUrl.ValueString(),
		ClientX509CertUrl:       data.ClientX509CertUrl.ValueString(),
	}, hasUnknown
}

func (data *FcmServiceAccountModel) ToFcmServiceAccountJsonStr() (string, error) {
	serviceAccount, hasUnknown := data.ToFcmServiceAccount()
	if hasUnknown {
		return "", errors.New("service account has unknown values")
	}
	serviceAccountJson, err := json.Marshal(serviceAccount)
	if err != nil {
		return "", err
	}
	return string(serviceAccountJson), nil
}
