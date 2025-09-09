package locality

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func SuppressSDKNullAssignment(k, old, new string, d *schema.ResourceData) bool {
	return (new == "" && old != "") || (new != "" && old == "")
}
