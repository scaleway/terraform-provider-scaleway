package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayAccountSSHKey() *schema.Resource {
	return dataSourceScalewayIamSSHKey()
}
