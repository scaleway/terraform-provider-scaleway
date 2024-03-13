package verify

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

func IsEmail() schema.SchemaValidateFunc {
	return func(v interface{}, key string) (warnings []string, errors []error) {
		email, isString := v.(string)
		if !isString {
			return nil, []error{fmt.Errorf("invalid email for key '%s': not a string", key)}
		}

		if !validation.IsEmail(email) {
			return nil, []error{fmt.Errorf("invalid email for key '%s': '%s': should contain valid '@' character", key, email)}
		}

		return
	}
}
