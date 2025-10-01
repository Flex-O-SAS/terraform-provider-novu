package customvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type exactlyOneSet struct {
	sibling string // e.g. "name"
}

func ExactlyOneSet(sibling string) exactlyOneSet {
	return exactlyOneSet{sibling: sibling}
}

func (m exactlyOneSet) Description(ctx context.Context) string {
	return "Forces exactly one attribute to be set between the attribute with the validator and the sibling attribute"
}
func (m exactlyOneSet) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m exactlyOneSet) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	parent := req.Path.ParentPath()
	sibPath := parent.AtName(m.sibling)

	var sibVal attr.Value
	diags := req.Config.GetAttribute(ctx, sibPath, &sibVal) // if missing, it’s “changed”
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	configSibSet := !sibVal.IsNull() && !sibVal.IsUnknown()
	attrSet := !req.ConfigValue.IsUnknown() && !req.ConfigValue.IsNull()

	if configSibSet == attrSet {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Exactly one attribute must be set between %s and %s", sibPath, req.Path.String()))
	}
}

func (m exactlyOneSet) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	parent := req.Path.ParentPath()
	sibPath := parent.AtName(m.sibling)

	var sibVal attr.Value
	_ = req.Config.GetAttribute(ctx, sibPath, &sibVal) // if missing, it’s “changed”
	configSibSet := !sibVal.IsNull() && !sibVal.IsUnknown()
	attrSet := !req.ConfigValue.IsUnknown() && !req.ConfigValue.IsNull()

	if configSibSet == attrSet {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Exactly one attribute must be set between %s and %s", sibPath, req.Path.String()))
	}
}
