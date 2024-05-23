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

func IsEmailList() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		list, ok := i.([]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("invalid type for key '%s': expected a list of strings", k))
			return warnings, errors
		}

		for _, li := range list {
			email, isString := li.(string)
			if !isString {
				errors = append(errors, fmt.Errorf("invalid type for key '%s': each item must be a string", k))
				continue
			}
			if _, err := IsEmail()(email, k); len(err) > 0 {
				errors = append(errors, err...)
			}
		}
		return warnings, errors
	}
}
