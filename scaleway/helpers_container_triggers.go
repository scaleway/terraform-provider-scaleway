package scaleway

import (
	"context"
	"time"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func waitForContainerTrigger(ctx context.Context, containerAPI *container.API, region scw.Region, id string, timeout time.Duration) (*container.Trigger, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	trigger, err := containerAPI.WaitForTrigger(&container.WaitForTriggerRequest{
		Region:        region,
		TriggerID:     id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return trigger, err
}

func expandContainerTriggerMnqSqsCreationConfig(i interface{}) *container.CreateTriggerRequestMnqSqsClientConfig {
	m := i.(map[string]interface{})

	mnqNamespaceID := expandID(m["namespace_id"].(string))

	req := &container.CreateTriggerRequestMnqSqsClientConfig{
		Queue:        m["queue"].(string),
		MnqProjectID: m["project_id"].(string),
		MnqRegion:    m["region"].(string),
	}

	if mnqNamespaceID != "" {
		req.MnqNamespaceID = &mnqNamespaceID
	}

	return req
}

func expandContainerTriggerMnqNatsCreationConfig(i interface{}) *container.CreateTriggerRequestMnqNatsClientConfig {
	m := i.(map[string]interface{})

	return &container.CreateTriggerRequestMnqNatsClientConfig{
		Subject:          m["subject"].(string),
		MnqProjectID:     m["project_id"].(string),
		MnqRegion:        m["region"].(string),
		MnqNatsAccountID: expandID(m["account_id"]),
	}
}

func completeContainerTriggerMnqCreationConfig(i interface{}, d *schema.ResourceData, meta interface{}, region scw.Region) error {
	m := i.(map[string]interface{})

	if sqsRegion, exists := m["region"]; !exists || sqsRegion == "" {
		m["region"] = region.String()
	}

	if projectID, exists := m["project_id"]; !exists || projectID == "" {
		projectID, _, err := extractProjectID(d, meta.(*Meta))
		if err == nil {
			m["project_id"] = projectID
		}
	}

	return nil
}
