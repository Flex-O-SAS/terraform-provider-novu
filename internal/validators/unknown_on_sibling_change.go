package customvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stringUnknownIfSiblingChanged struct {
	sibling string // e.g. "name"
}

func StringUnknownIfSiblingChanged(sibling string) planmodifier.String {
	return stringUnknownIfSiblingChanged{sibling: sibling}
}

func (m stringUnknownIfSiblingChanged) Description(ctx context.Context) string {
	return "Forces Unknown when the sibling attribute is changed. Useful to set known after apply inside a nested object in a list"
}
func (m stringUnknownIfSiblingChanged) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m stringUnknownIfSiblingChanged) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	parent := req.Path.ParentPath()
	sibPath := parent.AtName(m.sibling)

	// Read planned sibling
	var planSib types.String
	if diags := req.Plan.GetAttribute(ctx, sibPath, &planSib); diags.HasError() {
		// if we can’t read plan, be conservative
		resp.PlanValue = types.StringUnknown()
		return
	}

	// If sibling is unknown, propagate unknown
	if planSib.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	// Read prior state sibling
	var stateSib types.String
	_ = req.State.GetAttribute(ctx, sibPath, &stateSib) // if missing, it’s “changed”

	// If sibling changed (including was null before), mark this attr unknown
	if stateSib.IsNull() || stateSib.IsUnknown() || stateSib.ValueString() != planSib.ValueString() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	// Otherwise keep state to avoid churn
	if !req.StateValue.IsUnknown() && !req.StateValue.IsNull() {
		resp.PlanValue = req.StateValue
	}
}
