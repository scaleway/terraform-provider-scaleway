package scaleway

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func expandFunctionTriggerNatsCreationConfig(i interface{}) *function.CreateTriggerRequestMnqNatsClientConfig {
	mList := i.([]interface{})
	if len(mList) < 1 {
		return nil
	}
	m := mList[0].(map[string]interface{})

	return &function.CreateTriggerRequestMnqNatsClientConfig{
		MnqNamespaceID: m["namespace_id"].(string),
		Subject:        m["subject"].(string),
		MnqProjectID:   m["project_id"].(string),
		MnqRegion:      m["region"].(string),
	}
}

func expandFunctionTriggerMnqSqsCreationConfig(i interface{}) *function.CreateTriggerRequestMnqSqsClientConfig {
	m := i.(map[string]interface{})

	return &function.CreateTriggerRequestMnqSqsClientConfig{
		MnqNamespaceID: expandID(m["namespace_id"].(string)),
		Queue:          m["queue"].(string),
		MnqProjectID:   m["project_id"].(string),
		MnqRegion:      m["region"].(string),
	}
}

func completeFunctionTriggerMnqSqsCreationConfig(i interface{}, d *schema.ResourceData, meta interface{}, region scw.Region) error {
	m := i.(map[string]interface{})

	if sqsRegion, exists := m["region"]; !exists || sqsRegion == "" {
		m["region"] = region.String()
	}

	if projectID, exists := m["project_id"]; !exists || projectID == "" {
		projectID, _, err := extractProjectID(d, meta.(*Meta))
		if err != nil {
			return fmt.Errorf("failed to find a valid project_id")
		}
		m["project_id"] = projectID
	}

	return nil
}

func expandFunctionTriggerSqsCreationConfig(i interface{}) *function.CreateTriggerRequestSqsClientConfig {
	m := i.(map[string]interface{})

	return &function.CreateTriggerRequestSqsClientConfig{
		AccessKey: m["access_key"].(string),
		SecretKey: m["secret_key"].(string),
	}
}
