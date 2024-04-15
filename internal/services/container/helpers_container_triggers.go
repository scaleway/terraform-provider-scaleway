package container

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const defaultTriggerTimeout = 15 * time.Minute

func waitForContainerTrigger(ctx context.Context, containerAPI *container.API, region scw.Region, id string, timeout time.Duration) (*container.Trigger, error) {
	retryInterval := defaultTriggerRetryInterval
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
	req := &container.CreateTriggerRequestMnqSqsClientConfig{
		Queue:        m["queue"].(string),
		MnqProjectID: m["project_id"].(string),
		MnqRegion:    m["region"].(string),
	}

	return req
}

func expandContainerTriggerMnqNatsCreationConfig(i interface{}) *container.CreateTriggerRequestMnqNatsClientConfig {
	m := i.(map[string]interface{})

	return &container.CreateTriggerRequestMnqNatsClientConfig{
		Subject:          m["subject"].(string),
		MnqProjectID:     m["project_id"].(string),
		MnqRegion:        m["region"].(string),
		MnqNatsAccountID: locality.ExpandID(m["account_id"]),
	}
}

func completeContainerTriggerMnqCreationConfig(i interface{}, d *schema.ResourceData, m interface{}, region scw.Region) error {
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
