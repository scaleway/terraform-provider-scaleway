package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultContainerNamespaceTimeout = 5 * time.Minute
	defaultContainerCronTimeout      = 5 * time.Minute
	defaultContainerTimeout          = 5 * time.Minute
	defaultContainerRetryInterval    = 5 * time.Second
)

// containerAPIWithRegion returns a new container API and the region.
func containerAPIWithRegion(d *schema.ResourceData, m interface{}) (*container.API, scw.Region, error) {
	meta := m.(*Meta)
	api := container.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}
	return api, region, nil
}

// containerAPIWithRegionAndID returns a new container API, region and ID.
func containerAPIWithRegionAndID(m interface{}, id string) (*container.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := container.NewAPI(meta.scwClient)

	region, id, err := parseRegionalID(id)
	if err != nil {
		return nil, "", "", err
	}
	return api, region, id, nil
}

func setCreateContainerRequest(d *schema.ResourceData, region scw.Region) (*container.CreateContainerRequest, error) {
	// required
	nameRaw := d.Get("name")
	namespaceID := d.Get("namespace_id")

	name := expandOrGenerateString(nameRaw.(string), "co")
	privacyType := d.Get("privacy")
	protocol := d.Get("protocol")

	req := &container.CreateContainerRequest{
		Region:      region,
		NamespaceID: expandID(namespaceID),
		Name:        name,
		Privacy:     container.ContainerPrivacy(privacyType.(string)),
		Protocol:    container.ContainerProtocol(*expandStringPtr(protocol)),
	}

	// optional
	if envVariablesRaw, ok := d.GetOk("environment_variables"); ok {
		req.EnvironmentVariables = expandMapPtrStringString(envVariablesRaw)
	}

	if minScale, ok := d.GetOk("min_scale"); ok {
		req.MinScale = scw.Uint32Ptr(uint32(minScale.(int)))
	}

	if maxScale, ok := d.GetOk("max_scale"); ok {
		req.MaxScale = scw.Uint32Ptr(uint32(maxScale.(int)))
	}

	if memoryLimit, ok := d.GetOk("memory_limit"); ok {
		req.MemoryLimit = scw.Uint32Ptr(uint32(memoryLimit.(int)))
	}

	if timeout, ok := d.GetOk("timeout"); ok {
		timeInt := timeout.(int)
		req.Timeout = &scw.Duration{Seconds: int64(timeInt)}
	}

	if port, ok := d.GetOk("port"); ok {
		req.Port = scw.Uint32Ptr(uint32(port.(int)))
	}

	if description, ok := d.GetOk("description"); ok {
		req.Description = expandStringPtr(description)
	}

	if registryImage, ok := d.GetOk("registry_image"); ok {
		req.RegistryImage = expandStringPtr(registryImage)
	}

	if maxConcurrency, ok := d.GetOk("max_concurrency"); ok {
		req.MaxConcurrency = scw.Uint32Ptr(uint32(maxConcurrency.(int)))
	}

	return req, nil
}

func waitForContainerNamespace(ctx context.Context, containerAPI *container.API, region scw.Region, namespaceID string, timeout time.Duration) (*container.Namespace, error) {
	retryInterval := defaultContainerRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	ns, err := containerAPI.WaitForNamespace(&container.WaitForNamespaceRequest{
		Region:        region,
		NamespaceID:   namespaceID,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return ns, err
}

func waitForContainerCron(ctx context.Context, api *container.API, cronID string, region scw.Region, timeout time.Duration) (*container.Cron, error) {
	retryInterval := defaultContainerRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	request := container.WaitForCronRequest{
		CronID:        cronID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}

	return api.WaitForCron(&request, scw.WithContext(ctx))
}

func waitForContainer(ctx context.Context, api *container.API, containerID string, region scw.Region, timeout time.Duration) (*container.Container, error) {
	retryInterval := defaultContainerRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	request := container.WaitForContainerRequest{
		ContainerID:   containerID,
		Region:        region,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}

	return api.WaitForContainer(&request, scw.WithContext(ctx))
}

func expandContainerSecrets(secretsRawMap interface{}) []*container.Secret {
	secretsMap := secretsRawMap.(map[string]interface{})
	secrets := make([]*container.Secret, 0, len(secretsMap))

	for k, v := range secretsMap {
		secrets = append(secrets, &container.Secret{
			Key:   k,
			Value: expandStringPtr(v),
		})
	}

	return secrets
}
