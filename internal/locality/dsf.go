package locality

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// SuppressSDKNullAssignment aims to remove apply inconsistency on SDKv2 based resources
func SuppressSDKNullAssignment(k, old, new string, d *schema.ResourceData) bool {
	return (new == "" && old != "") || (new != "" && old == "")
}
