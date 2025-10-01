package helpers

import (
	"fmt"
	"strings"
)

func DisplayBetterSlugError(err error, otherFields ...string) string {
	if err == nil {
		return ""
	}
	otherFields = append(otherFields, "slug")
	otherFieldsString := strings.Join(otherFields, " or ")

	return strings.ReplaceAll(err.Error(), "must be a valid slug format", fmt.Sprintf("must be a valid %s format", otherFieldsString))
}
