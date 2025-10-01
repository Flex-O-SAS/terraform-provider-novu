package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StringUnkownOnSiblingNotKnownVal struct {
	sibling string
}

//nolint:unparam // kept generic for future attributes;
func stringUnkownOnSiblingNotKnownVal(sibling string) planmodifier.String {
	return StringUnkownOnSiblingNotKnownVal{sibling: sibling}
}

func (m StringUnkownOnSiblingNotKnownVal) Description(ctx context.Context) string {
	return "Forces Unknown when the sibling attribute is nil. Useful to set known after apply inside a nested object in a list"
}
func (m StringUnkownOnSiblingNotKnownVal) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m StringUnkownOnSiblingNotKnownVal) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If already decided (Replace, etc.), don’t fight it
	if !resp.PlanValue.IsUnknown() && !resp.PlanValue.IsNull() {
		return
	}

	// Find the sibling id attribute on the parent element
	parent := req.Path.ParentPath() // ...steps.<index>
	idPath := parent.AtName(m.sibling)

	var stateID types.String
	// If we can't read it, treat as new
	if diags := req.State.GetAttribute(ctx, idPath, &stateID); diags.HasError() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	if stateID.IsNull() || stateID.IsUnknown() || stateID.ValueString() == "" {
		// New element ⇒ make computed field Unknown so it’s filled on Read
		resp.PlanValue = types.StringUnknown()
		return
	}

	// Existing element ⇒ prefer state (no churn)
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

type objectUnkownOnSiblingNotKnownVal struct {
	sibling string
}

func ObjectUnkownOnSiblingNotKnownVal(sibling string) planmodifier.Object {
	return objectUnkownOnSiblingNotKnownVal{sibling: sibling}
}

func (m objectUnkownOnSiblingNotKnownVal) Description(ctx context.Context) string {
	return "Forces Unknown when the sibling attribute is nil. Useful to set known after apply inside a nested object in a list"
}
func (m objectUnkownOnSiblingNotKnownVal) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m objectUnkownOnSiblingNotKnownVal) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// If already decided (Replace, etc.), don’t fight it
	if !resp.PlanValue.IsUnknown() && !resp.PlanValue.IsNull() {
		return
	}

	// Find the sibling id attribute on the parent element
	parent := req.Path.ParentPath() // ...steps.<index>
	idPath := parent.AtName(m.sibling)

	var stateID types.String
	// If we can't read it, treat as new
	if diags := req.State.GetAttribute(ctx, idPath, &stateID); diags.HasError() {

		// Derive the ObjectType from plan (or fall back to state)
		t := req.PlanValue.Type(ctx)
		ot, ok := t.(types.ObjectType)
		if !ok {
			t = req.StateValue.Type(ctx)
			ot, ok = t.(types.ObjectType)
			if !ok {
				// don't throw, just return
				resp.Diagnostics.AddWarning("Invalid ObjectType", "Could not derive the ObjectType from the plan or state")
				return
			}
		}
		resp.PlanValue = types.ObjectUnknown(ot.AttrTypes)
		return
	}
}
