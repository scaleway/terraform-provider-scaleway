package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceScalewayAccountSSKKey() *schema.Resource {
	return resourceScalewayIamSSKKey()
}
