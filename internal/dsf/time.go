package dsf

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Duration(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	d1, err1 := time.ParseDuration(oldValue)
	d2, err2 := time.ParseDuration(newValue)
	if err1 != nil || err2 != nil {
		return false
	}

	return d1 == d2
}

func TimeRFC3339(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	t1, err1 := time.Parse(time.RFC3339, oldValue)
	t2, err2 := time.Parse(time.RFC3339, newValue)
	if err1 != nil || err2 != nil {
		return false
	}

	return t1.Equal(t2)
}
