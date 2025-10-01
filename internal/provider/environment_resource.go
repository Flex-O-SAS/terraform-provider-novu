package provider

import (
	"context"
	"errors"
	"fmt"
	"terraform-provider-novu/internal/helpers"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	novugo "github.com/novuhq/novu-go"
	"github.com/novuhq/novu-go/models/apierrors"
	"github.com/novuhq/novu-go/models/components"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	defaultColor = "#3498db"

	_ resource.Resource               = &EnvironmentResource{}
	_ resource.ResourceWithModifyPlan = &EnvironmentResource{}
	// _ resource.ResourceWithImportState = &EnvironmentResource{}
)

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

// ExampleDataSource defines the data source implementation.
type EnvironmentResource struct {
	client *novugo.Novu
}

type EnvironmentResourceModel struct {
	Id    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
	// OrganizationId types.String            `tfsdk:"organization_id"` // Useless
	Identifier   types.String `tfsdk:"identifier"`
	ParentId     types.String `tfsdk:"parent_environment_id"`
	Slug         types.String `tfsdk:"slug"`
	IsProduction types.Bool   `tfsdk:"is_production"`
}
type ApiKeyResourceModel struct {
	Value   types.String `tfsdk:"value"`
	OwnerId types.String `tfsdk:"owner_id"`
	Hash    types.String `tfsdk:"hash"`
}

func (d *EnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environment resource. Manage any environment within the organization. \n" +
			"WARNING : this resource could not be properly tested ! \n" +
			"In the free and Pro plans, only the color can be updated, everything else throws an error. \n" +
			"If you are in one of these plans, Prefer using the data source instead. ",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The environment id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The environment name",
				Required:            true,
			},
			// "organization_id": schema.StringAttribute{
			// 	MarkdownDescription: "The environment's organization id",
			// 	Computed:            true,
			// },
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The environment identifier. Can be updated (on the Team and Enterprise plans) but not set at creation.",
				Optional:            true, // can be updated but not set at creation.
				Computed:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The environment color. The default color is %s. This value can not be read from the API, so changes in the interface will be ignored.", defaultColor),
				Optional:            true, // because not returned by the API.
				Computed:            true, // Default color is set in importState because it's mandatory at creation
			},
			"parent_environment_id": schema.StringAttribute{
				MarkdownDescription: "The environment's parent id. Not sure if it requires replace. Tested not editable on the free and Pro plans.",
				Optional:            true,
				Computed:            true, // in case of import we don't have to set this
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
	}
}

func handleColor(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var configColor types.String
	diags := req.Config.GetAttribute(ctx, path.Root("color"), &configColor)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// color is set in the config, we do nothing
	if !configColor.IsNull() {
		return
	}

	if req.State.Raw.IsNull() {
		// it is creation, we set a default value
		diags := resp.Plan.SetAttribute(ctx, path.Root("color"), types.StringValue(defaultColor))
		resp.Diagnostics.Append(diags...)
	} else {
		// not a creation we make the null value known in the plan
		diags := resp.Plan.SetAttribute(ctx, path.Root("color"), types.StringNull())
		resp.Diagnostics.Append(diags...)

	}
}

func markKnownIfNoNameChange(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, fieldsToMarkAsKnown []string) {
	var configName types.String
	var stateName types.String
	var planName types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("name"), &configName)
	resp.Diagnostics.Append(diags...)
	diags = req.State.GetAttribute(ctx, path.Root("name"), &stateName)
	resp.Diagnostics.Append(diags...)
	diags = req.Plan.GetAttribute(ctx, path.Root("name"), &planName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// there's a name change, we don't mark fields as known
	if !planName.Equal(stateName) {
		return
	}

	for _, field := range fieldsToMarkAsKnown {
		var configField types.String
		var stateField types.String
		diags = req.Config.GetAttribute(ctx, path.Root(field), &configField)
		resp.Diagnostics.Append(diags...)
		diags = req.State.GetAttribute(ctx, path.Root(field), &stateField)
		resp.Diagnostics.Append(diags...)
		// don't mark unknown if the field is set or changed
		if !configField.IsNull() && !configField.Equal(stateField) {
			tflog.Info(ctx, "field not marked as known", map[string]interface{}{
				"field":      field,
				"is_null":    configField.IsNull(),
				"is_unknown": configField.IsUnknown(),
				"is_equal":   configField.Equal(stateField),
			})
			continue
		}
		diags = resp.Plan.SetAttribute(ctx, path.Root(field), stateField)
		resp.Diagnostics.Append(diags...)
	}
}

func (d *EnvironmentResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// it's a deletion, we don't care
	if req.Plan.Raw.IsNull() {
		return
	}

	handleColor(ctx, req, resp)

	markKnownIfNoNameChange(ctx, req, resp, []string{"slug", "identifier"})

	// handle is_production
	var parent_id types.String
	diags := resp.Plan.GetAttribute(ctx, path.Root("parent_environment_id"), &parent_id)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// parent_id is unknown at plan (ex: from a variable), we don't make a guess
	if !parent_id.IsUnknown() {
		resp.Plan.SetAttribute(ctx, path.Root("is_production"), types.BoolValue(!parent_id.IsNull()))
	}
}

func (d *EnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *novugo.Novu or *ProviderClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
}

func (d *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	if data.Name.IsNull() || data.Name.ValueString() == "" {
		resp.Diagnostics.AddError("Client Error", "Name is required")
	}
	if data.Color.IsNull() || data.Color.IsUnknown() || data.Color.ValueString() == "" {
		resp.Diagnostics.AddError("Client Error", "Color is required")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the create request
	createReq := components.CreateEnvironmentRequestDto{
		Name:  data.Name.ValueString(),
		Color: data.Color.ValueString(),
	}

	// Only set ParentID if it's not null/empty
	if !data.ParentId.IsNull() && !data.ParentId.IsUnknown() && data.ParentId.ValueString() != "" {
		createReq.ParentID = novugo.String(data.ParentId.ValueString())
	}

	// Add logging for debugging
	tflog.Info(ctx, "Creating environment", map[string]interface{}{
		"name":      createReq.Name,
		"color":     createReq.Color,
		"parent_id": createReq.ParentID,
	})

	res, err := d.client.Environments.Create(ctx, createReq, nil)

	if err != nil {
		tflog.Info(ctx, "Detailed error", map[string]interface{}{
			"detailed_error": err.Error(),
		})
		var e1 *apierrors.ErrorDto
		if errors.As(err, &e1) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Detailed error: %s \n"+
				"Are you sure you are on the right Novu plan ? ", e1.Message))
		}
		var e2 *apierrors.ValidationErrorDto
		if errors.As(err, &e2) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Detailed error: %s \n"+
				"Are you sure you are on the right Novu plan ? ", e2.Message))
		}
		var e3 *apierrors.APIError
		if errors.As(err, &e3) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Detailed error: %s \n"+
				"Are you sure you are on the right Novu plan ? ", e3.Message))
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment: %s \n"+
			"Are you sure you are on the right Novu plan ? ", err))
		return
	}

	if res == nil || res.EnvironmentResponseDto == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment: %s", "Response is nil"))
		return
	}

	// set computed and optional attributes
	// we update all the values to be sure the state is correct
	if res.EnvironmentResponseDto.ParentID != nil {
		data.ParentId = types.StringValue(*res.EnvironmentResponseDto.ParentID)
		data.IsProduction = types.BoolValue(true)
	} else {
		data.ParentId = types.StringNull()
		data.IsProduction = types.BoolValue(false)
	}
	if res.EnvironmentResponseDto.Slug != nil {
		data.Slug = types.StringValue(*res.EnvironmentResponseDto.Slug)
	} else {
		data.Slug = types.StringNull()
	}
	data.Name = types.StringValue(res.EnvironmentResponseDto.Name)
	data.Id = types.StringValue(res.EnvironmentResponseDto.ID)
	data.Identifier = types.StringValue(res.EnvironmentResponseDto.Identifier)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnvironmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := d.client.Environments.Delete(ctx, data.Id.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete environment: %s", err))
		return
	}
}

func (d *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnvironmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Environments.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read environment: %s", err))
		return
	}

	var foundEnvironment *components.EnvironmentResponseDto
	environmentsList := res.EnvironmentResponseDtos

	for i := range environmentsList {
		// search by id if set
		if !data.Id.IsNull() && environmentsList[i].ID == data.Id.ValueString() {
			foundEnvironment = &environmentsList[i]
			break
		}
		// search by name if id is not set
		if data.Id.IsNull() && environmentsList[i].Name == data.Name.ValueString() {
			foundEnvironment = &environmentsList[i]
			break
		}
	}

	if foundEnvironment != nil {
		if foundEnvironment.ParentID != nil {
			data.ParentId = types.StringValue(*foundEnvironment.ParentID)
			data.IsProduction = types.BoolValue(true)
		} else {
			data.ParentId = types.StringNull()
			data.IsProduction = types.BoolValue(false)
		}
		data.Id = types.StringValue(foundEnvironment.ID)
		data.Name = types.StringValue(foundEnvironment.Name)
		//data.Color = types.StringValue(foundEnvironment.Color) // this is not returned by the API
		data.Identifier = types.StringValue(foundEnvironment.Identifier)

		data.Slug = types.StringValue(*foundEnvironment.Slug)
	} else {
		errorMsg := "Environment"
		if data.Id.ValueString() != "" {
			errorMsg += " with id " + data.Id.ValueString()
		}
		if data.Name.ValueString() != "" {
			errorMsg += " with name " + data.Name.ValueString()
		}
		if data.Identifier.ValueString() != "" {
			errorMsg += " with identifier " + data.Identifier.ValueString()
		}
		errorMsg += " not found"
		resp.Diagnostics.AddError("Client Error", errorMsg)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentResourceModel

	var state EnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var updateReq components.UpdateEnvironmentRequestDto

	helpers.SetStringPtrIfChanged(plan.ParentId, state.ParentId, &updateReq.ParentID)
	helpers.SetStringPtrIfChanged(plan.Name, state.Name, &updateReq.Name)
	helpers.SetStringPtrIfChanged(plan.Identifier, state.Identifier, &updateReq.Identifier)
	helpers.SetStringPtrIfChanged(plan.Color, state.Color, &updateReq.Color)

	// no changes
	if updateReq.ParentID == nil && updateReq.Name == nil && updateReq.Identifier == nil && updateReq.Color == nil {
		return
	}

	if plan.Id.IsNull() || plan.Id.IsUnknown() {
		resp.Diagnostics.AddError("Client error", "Unable to read environment ID")
		return
	}

	name := updateReq.Name

	tflog.Info(ctx, "le foutu name", map[string]interface{}{
		"name": name,
	})

	tflog.Info(ctx, "Updating environment", map[string]interface{}{
		"id":         plan.Id.ValueString(),
		"update_req": updateReq,
	})

	httpRes, err := r.client.Environments.Update(ctx, plan.Id.ValueString(), updateReq, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment: %s"+
			"Are you sure you are on the right Novu plan and this environment can be updated? ", err))
		return
	}
	if httpRes.EnvironmentResponseDto == nil {
		resp.Diagnostics.AddError("Client Error", "The API did not return the expected data, impossible to determine if the update was successful")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}
