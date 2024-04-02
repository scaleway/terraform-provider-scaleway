package verify

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// IsDate will validate that field is a valid ISO 8601
// It is the same as RFC3339
func IsDate() schema.SchemaValidateDiagFunc {
	return func(i interface{}, _ cty.Path) diag.Diagnostics {
		date, isStr := i.(string)
		if !isStr {
			return diag.Errorf("%v is not a string", date)
		}
		_, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}
}

func IsDuration() schema.SchemaValidateFunc {
	return func(i interface{}, _ string) (strings []string, errors []error) {
		str, isStr := i.(string)
		if !isStr {
			return nil, []error{fmt.Errorf("%v is not a string", i)}
		}
		_, err := time.ParseDuration(str)
		if err != nil {
			return nil, []error{fmt.Errorf("cannot parse duration for value %s", str)}
		}
		return nil, nil
	}
}
