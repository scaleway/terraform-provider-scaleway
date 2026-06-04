package container

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/container/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultTriggerTimeout       = 15 * time.Minute
	defaultTriggerRetryInterval = 5 * time.Second
)

func expandDestinationConfig(destConfRaw any) (*container.CreateTriggerRequestDestinationConfig, error) {
	destConfMap := destConfRaw.(map[string]any)
	destinationConfigReq := &container.CreateTriggerRequestDestinationConfig{
		HTTPPath: destConfMap["http_path"].(string),
	}

	switch destConfMap["http_method"].(string) {
	case "get":
		destinationConfigReq.HTTPMethod = container.CreateTriggerRequestDestinationConfigHTTPMethodGet
	case "patch":
		destinationConfigReq.HTTPMethod = container.CreateTriggerRequestDestinationConfigHTTPMethodPatch
	case "post":
		destinationConfigReq.HTTPMethod = container.CreateTriggerRequestDestinationConfigHTTPMethodPost
	case "put":
		destinationConfigReq.HTTPMethod = container.CreateTriggerRequestDestinationConfigHTTPMethodPut
	case "delete":
		destinationConfigReq.HTTPMethod = container.CreateTriggerRequestDestinationConfigHTTPMethodDelete
	default:
		return nil, fmt.Errorf("unhandled HTTP method: %s", destConfMap["http_method"].(string))
	}

	return destinationConfigReq, nil
}

func expandContainerTriggerSqsCreationConfig(i any, region scw.Region) *container.CreateTriggerRequestSQSConfig {
	sqsConfig := i.(map[string]any)

	var regionToSet scw.Region
	if sqsRegion, exists := sqsConfig["region"]; !exists || sqsRegion == "" {
		regionToSet = region
	} else {
		regionToSet = scw.Region(sqsRegion.(string))
	}

	req := &container.CreateTriggerRequestSQSConfig{
		Region:          regionToSet,
		Endpoint:        sqsConfig["endpoint"].(string),
		AccessKeyID:     sqsConfig["access_key"].(string),
		SecretAccessKey: sqsConfig["secret_key"].(string),
		QueueURL:        sqsConfig["queue_url"].(string),
	}

	return req
}

func expandContainerTriggerNatsCreationConfig(i any) *container.CreateTriggerRequestNATSConfig {
	natsConfig := i.(map[string]any)

	req := &container.CreateTriggerRequestNATSConfig{
		CredentialsFileContent: natsConfig["credentials_file_content"].(string),
		ServerURLs:             types.ExpandStrings(natsConfig["server_urls"]),
		Subject:                natsConfig["subject"].(string),
	}

	return req
}

func expandContainerTriggerCronCreationConfig(i any) *container.CreateTriggerRequestCronConfig {
	cronConfig := i.(map[string]any)
	headersMapInterface := cronConfig["headers"].(map[string]any)
	headersMapString := make(map[string]string, len(headersMapInterface))

	for key, valueI := range headersMapInterface {
		headersMapString[key] = valueI.(string)
	}

	return &container.CreateTriggerRequestCronConfig{
		Schedule: cronConfig["schedule"].(string),
		Timezone: cronConfig["timezone"].(string),
		Body:     cronConfig["body"].(string),
		Headers:  headersMapString,
	}
}

func flattenDestinationConfig(destinationConfig *container.TriggerDestinationConfig) any {
	destinationConfigFlat := make([]map[string]any, 0, 1)

	return append(destinationConfigFlat, map[string]any{
		"http_path":   destinationConfig.HTTPPath,
		"http_method": destinationConfig.HTTPMethod.String(),
	})
}

func flattenTriggerSqs(d *schema.ResourceData, sqsConfig *container.TriggerSQSConfig) any {
	sqsConfigFlat := make([]map[string]any, 0, 1)

	if sqsConfig == nil {
		return sqsConfigFlat
	}

	sqsConfigFlat = append(sqsConfigFlat, map[string]any{
		"endpoint":  sqsConfig.Endpoint,
		"queue_url": sqsConfig.QueueURL,
	})

	// Retrieve from the state attributes that were stored at creation because only available there
	if state := d.State(); state != nil {
		if accessKey, ok := state.Attributes["sqs.0.access_key"]; ok {
			sqsConfigFlat[0]["access_key"] = accessKey
		}

		if secretKey, ok := state.Attributes["sqs.0.secret_key"]; ok {
			sqsConfigFlat[0]["secret_key"] = secretKey
		}
	}

	return sqsConfigFlat
}

func flattenTriggerNats(d *schema.ResourceData, natsConfig *container.TriggerNATSConfig) any {
	natsConfigFlat := make([]map[string]any, 0, 1)

	if natsConfig == nil {
		return natsConfigFlat
	}

	natsConfigFlat = append(natsConfigFlat, map[string]any{
		"subject":     natsConfig.Subject,
		"server_urls": natsConfig.ServerURLs,
	})

	// Retrieve from the state attributes that were stored at creation because only available there
	if state := d.State(); state != nil {
		if credentialsFileContent, ok := state.Attributes["nats.0.credentials_file_content"]; ok {
			natsConfigFlat[0]["credentials_file_content"] = credentialsFileContent
		}
	}

	return natsConfigFlat
}

func flattenTriggerCron(cronConfig *container.TriggerCronConfig) any {
	cronConfigFlat := make([]map[string]any, 0, 1)

	if cronConfig == nil {
		return cronConfigFlat
	}

	cronConfigFlat = append(cronConfigFlat, map[string]any{
		"schedule": cronConfig.Schedule,
		"timezone": cronConfig.Timezone,
		"body":     cronConfig.Body,
		"headers":  types.FlattenMap(cronConfig.Headers),
	})

	return cronConfigFlat
}

func updateDestinationConfig(destConfRaw any) (*container.UpdateTriggerRequestDestinationConfig, error) {
	destConfMap := destConfRaw.(map[string]any)
	destinationConfigReq := &container.UpdateTriggerRequestDestinationConfig{
		HTTPPath: types.ExpandUpdatedStringPtr(destConfMap["http_path"]),
	}

	switch destConfMap["http_method"].(string) {
	case "get":
		destinationConfigReq.HTTPMethod = new(container.UpdateTriggerRequestDestinationConfigHTTPMethodGet)
	case "patch":
		destinationConfigReq.HTTPMethod = new(container.UpdateTriggerRequestDestinationConfigHTTPMethodPatch)
	case "post":
		destinationConfigReq.HTTPMethod = new(container.UpdateTriggerRequestDestinationConfigHTTPMethodPost)
	case "put":
		destinationConfigReq.HTTPMethod = new(container.UpdateTriggerRequestDestinationConfigHTTPMethodPut)
	case "delete":
		destinationConfigReq.HTTPMethod = new(container.UpdateTriggerRequestDestinationConfigHTTPMethodDelete)
	default:
		return nil, fmt.Errorf("unhandled HTTP method: %s", destConfMap["http_method"].(string))
	}

	return destinationConfigReq, nil
}

func updateSqsConfig(sqs any) *container.UpdateTriggerRequestSQSConfig {
	sqsFlat := sqs.(map[string]any)

	req := &container.UpdateTriggerRequestSQSConfig{
		QueueURL:        types.ExpandUpdatedStringPtr(sqsFlat["queue_url"]),
		Endpoint:        types.ExpandUpdatedStringPtr(sqsFlat["endpoint"]),
		AccessKeyID:     types.ExpandUpdatedStringPtr(sqsFlat["access_key"]),
		SecretAccessKey: types.ExpandUpdatedStringPtr(sqsFlat["secret_key"]),
	}

	return req
}

func updateNatsConfig(nats any) *container.UpdateTriggerRequestNATSConfig {
	natsFlat := nats.(map[string]any)

	reqI := &container.UpdateTriggerRequestNATSConfig{
		ServerURLs:             types.ExpandUpdatedStringsPtr(natsFlat["server_urls"]),
		Subject:                types.ExpandUpdatedStringPtr(natsFlat["subject"]),
		CredentialsFileContent: types.ExpandUpdatedStringPtr(natsFlat["credentials_file_content"]),
	}

	return reqI
}

func updateCronConfig(cron any) *container.UpdateTriggerRequestCronConfig {
	cronFlat := cron.(map[string]any)

	req := &container.UpdateTriggerRequestCronConfig{
		Schedule: types.ExpandUpdatedStringPtr(cronFlat["schedule"]),
		Timezone: types.ExpandUpdatedStringPtr(cronFlat["timezone"]),
		Body:     types.ExpandUpdatedStringPtr(cronFlat["body"]),
		Headers:  types.ExpandMapPtrStringString(cronFlat["headers"]),
	}

	return req
}

// forceNewOnSourceChange triggers a ForceNew if the source of the trigger changes as this update is not allowed by the API
func forceNewOnSourceChange(keys ...string) schema.CustomizeDiffFunc {
	return func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
		for _, key := range keys {
			oldConfig, newConfig := diff.GetChange(key)
			if len(oldConfig.([]any)) > len(newConfig.([]any)) {
				return diff.ForceNew(key)
			}
		}

		return nil
	}
}
