package verify

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/robfig/cron/v3"
)

func ValidateCronExpression() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of '%s' to be string", k))
			return
		}
		_, err := cron.ParseStandard(v)
		if err != nil {
			es = append(es, fmt.Errorf("'%s' should be an valid Cron expression", k))
		}
		return
	}
}
