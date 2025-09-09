package locality

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// SuppressSDKNullAssignment aims to remove apply inconsistency on SDKv2 based resources
func SuppressSDKNullAssignment(k, oldValue, newValue string, d *schema.ResourceData) bool {
	return (newValue == "" && oldValue != "") || (newValue != "" && oldValue == "")
}
