package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func expandFunctionTriggerMnqSqsCreationConfig(i interface{}) *function.CreateTriggerRequestMnqSqsClientConfig {
	m := i.(map[string]interface{})

	mnqNamespaceID := locality.ExpandID(m["namespace_id"].(string))

	req := &function.CreateTriggerRequestMnqSqsClientConfig{
		Queue:        m["queue"].(string),
		MnqProjectID: m["project_id"].(string),
		MnqRegion:    m["region"].(string),
	}

	if mnqNamespaceID != "" {
		req.MnqNamespaceID = &mnqNamespaceID
	}

	return req
}

func expandFunctionTriggerMnqNatsCreationConfig(i interface{}) *function.CreateTriggerRequestMnqNatsClientConfig {
	m := i.(map[string]interface{})

	return &function.CreateTriggerRequestMnqNatsClientConfig{
		Subject:          locality.ExpandID(m["subject"]),
		MnqProjectID:     m["project_id"].(string),
		MnqRegion:        m["region"].(string),
		MnqNatsAccountID: locality.ExpandID(m["account_id"]),
	}
}

func completeFunctionTriggerMnqCreationConfig(i interface{}, d *schema.ResourceData, m interface{}, region scw.Region) error {
	configMap := i.(map[string]interface{})

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
