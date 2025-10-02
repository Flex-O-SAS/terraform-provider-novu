package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-novu/internal/helpers"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	novugo "github.com/novuhq/novu-go"
	"github.com/novuhq/novu-go/models/components"
)

var (
	_ resource.Resource              = &TopicResource{}
	_ resource.ResourceWithConfigure = &TopicResource{}
)

func NewTopicResource() resource.Resource {
	return &TopicResource{}
}

type TopicResource struct {
	client *novugo.Novu
}

type TopicResourceModel struct {
	Key  types.String `tfsdk:"key"`
	Name types.String `tfsdk:"name"`
}

func (r *TopicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topic"
}

func (r *TopicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if clients, ok := req.ProviderData.(*ProviderClients); ok && clients.Novu != nil {
		r.client = clients.Novu
	} else if client, ok := req.ProviderData.(*novugo.Novu); ok {
		r.client = client
	} else {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *ProviderClients or *novugo.Novu, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
}

func (r *TopicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Topic resource",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "Key of the topic",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^(?:[A-Za-z0-9:_-]+|[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,})$`),
						"The key must contain only alphanumeric characters (a-z, A-Z, 0-9), hyphens (-), underscores (_), colons (:), or be a valid email address",
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the topic",
				Optional:            true,
			},
		},
	}
}

func (r *TopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config TopicResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := helpers.FromTfString(config.Key)
	name := helpers.FromTfString(config.Name)

	if key == nil {
		resp.Diagnostics.AddError("Client error", "Key is required")
		return
	}

	createReq := components.CreateUpdateTopicRequestDto{
		Key:  *key,
		Name: name,
	}

	res, err := r.client.Topics.Create(ctx, createReq, nil)
	if err != nil {
		errMsg := err.Error()
		// SDK returns response body closed error if the key contains invalid characters
		if strings.Contains(errMsg, "response body closed") {
			err = errors.Join(err, errors.New("it may be that the key contains invalid characters"))
		}
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Unable to create topic, got error: %s", err))
		return
	}

	if res == nil || res.TopicResponseDto == nil {
		resp.Diagnostics.AddError("Client error", "Unable to create topic, got error: response is nil")
		return
	}

	config.Key = helpers.TfString(&res.TopicResponseDto.Key)
	config.Name = helpers.TfString(res.TopicResponseDto.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (r *TopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := helpers.FromTfString(state.Key)
	if key == nil {
		resp.Diagnostics.AddError("Client error", "Key is required")
		return
	}

	res, err := r.client.Topics.Get(ctx, *key, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Unable to read topic, got error: %s", err))
		return
	}

	if res == nil || res.TopicResponseDto == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Key = helpers.TfString(&res.TopicResponseDto.Key)
	state.Name = helpers.TfString(res.TopicResponseDto.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TopicResourceModel
	var state TopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Key.IsNull() || plan.Key.IsUnknown() {
		resp.Diagnostics.AddError("Client error", "Key is required but was not set")
		return
	}

	if plan.Name.IsUnknown() {
		resp.Diagnostics.AddError("Client error", "Name is unknown during update")
		return
	}

	if plan.Name.IsNull() {
		resp.Diagnostics.AddError("Client error", "Name is required but was not set")
		return
	}

	updateReq := components.UpdateTopicRequestDto{
		Name: plan.Name.ValueString(),
	}

	res, err := r.client.Topics.Update(ctx, *helpers.FromTfString(plan.Key), updateReq, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Unable to update topic, got error: %s", err))
		return
	}

	if res == nil || res.TopicResponseDto == nil {
		resp.Diagnostics.AddError("Client error", "Unable to update topic, got error: response is nil")
		return
	}

	state.Name = helpers.TfString(res.TopicResponseDto.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Key.IsNull() || state.Key.IsUnknown() {
		resp.Diagnostics.AddError("Client error", "Key is required but was not set")
		return
	}

	_, err := r.client.Topics.Delete(ctx, state.Key.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Unable to delete topic, got error: %s", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *TopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
