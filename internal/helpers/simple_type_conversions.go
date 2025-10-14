package helpers

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Ptr[T any](v T) *T { return &v }

func TfString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func TfBool(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}

func FromTfBool(b types.Bool) *bool {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	return Ptr(b.ValueBool())
}

func FromTfString(s types.String) *string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	return Ptr(s.ValueString())
}

// Set field to the plan value if it is not null or unknown and if it is different from the state value
func SetStringPtrIfChanged(planVal, stateVal types.String, field **string) {
	if stateVal.Equal(planVal) || planVal.IsUnknown() || planVal.IsNull() {
		return
	}
	v := planVal.ValueString()
	*field = &v
}

// Set field to the plan value if it is not null or unknown and if it is different from the state value
func SetBoolPtrIfChanged(planVal, stateVal types.Bool, field **bool) {
	if stateVal.Equal(planVal) || planVal.IsUnknown() || planVal.IsNull() {
		return
	}
	v := planVal.ValueBool()
	*field = &v
}
