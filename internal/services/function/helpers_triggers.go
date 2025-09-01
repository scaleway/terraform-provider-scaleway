package function

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func expandFunctionTriggerMnqSqsCreationConfig(i any) *function.CreateTriggerRequestMnqSqsClientConfig {
	m := i.(map[string]any)
	req := &function.CreateTriggerRequestMnqSqsClientConfig{
		Queue:        m["queue"].(string),
		MnqProjectID: m["project_id"].(string),
		MnqRegion:    m["region"].(string),
	}

	return req
}

func expandFunctionTriggerMnqNatsCreationConfig(i any) *function.CreateTriggerRequestMnqNatsClientConfig {
	m := i.(map[string]any)

	return &function.CreateTriggerRequestMnqNatsClientConfig{
		Subject:          locality.ExpandID(m["subject"]),
		MnqProjectID:     m["project_id"].(string),
		MnqRegion:        m["region"].(string),
		MnqNatsAccountID: locality.ExpandID(m["account_id"]),
	}
}

func completeFunctionTriggerMnqCreationConfig(i any, d *schema.ResourceData, m any, region scw.Region) error {
	configMap := i.(map[string]any)

	if sqsRegion, exists := configMap["region"]; !exists || sqsRegion == "" {
		configMap["region"] = region.String()
	}

	if projectID, exists := configMap["project_id"]; !exists || projectID == "" {
		projectID, _, err := meta.ExtractProjectID(d, m)
		if err == nil {
			configMap["project_id"] = projectID
		}
	}

	return nil
}
