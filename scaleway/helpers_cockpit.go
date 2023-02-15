package scaleway

import (
	"context"
	"time"

	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultCockpitTimeout = 5 * time.Minute
)

// cockpitAPI returns a new container cockpit API.
func cockpitAPI(m interface{}) (*cockpit.API, error) {
	meta := m.(*Meta)
	api := cockpit.NewAPI(meta.scwClient)

	return api, nil
}

func getCockpitActivated(api *cockpit.API, projectId string) (*cockpit.Cockpit, error) {
	res, err := api.GetCockpit(&cockpit.GetCockpitRequest{
		ProjectID: projectId,
	})
	if err != nil {
		_, err := api.ActivateCockpit(&cockpit.ActivateCockpitRequest{
			ProjectID: projectId,
		})
		if err != nil {
			return nil, err
		}

		res, err = api.GetCockpit(&cockpit.GetCockpitRequest{
			ProjectID: projectId,
		})
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func flattenCockpitEndpoints(endpoints *cockpit.CockpitEndpoints) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"metrics_url":      endpoints.MetricsURL,
			"logs_url":         endpoints.LogsURL,
			"alertmanager_url": endpoints.AlertmanagerURL,
			"grafana_url":      endpoints.GrafanaURL,
		},
	}
}

func waitForCockpit(ctx context.Context, api *cockpit.API, projectID string, timeout time.Duration) (*cockpit.Cockpit, error) {
	retryInterval := defaultContainerRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	return api.WaitForCockpit(&cockpit.WaitForCockpitRequest{
		ProjectID:     projectID,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}
