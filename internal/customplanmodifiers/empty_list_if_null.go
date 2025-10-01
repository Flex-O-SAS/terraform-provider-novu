package customplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type emptyListIfNull struct {
}

func EmptyListIfNull() planmodifier.List {
	return emptyListIfNull{}
}

func (e emptyListIfNull) Description(ctx context.Context) string {
	return "Sets as empty list if null"
}

func (e emptyListIfNull) MarkdownDescription(ctx context.Context) string {
	return e.Description(ctx)
}

func (e emptyListIfNull) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Only act when user set nothing (config is null). If unknown, leave it.
	if req.ConfigValue.IsNull() {
		empty, diags := types.ListValue(req.ConfigValue.ElementType(ctx), []attr.Value{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.PlanValue = empty
	}
}
